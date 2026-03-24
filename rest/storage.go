// GCS Google Cloud Storage API
package rest

import (
	"context"
	"io"
	"io/fs"
	"os"
	ppath "path"
	"path/filepath"
	"runtime"
	"strings"

	"log"

	"cloud.google.com/go/storage"
	"github.com/tonymet/gcloud-go/throttle"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type AuthorizedStorageClient struct {
	sClient *storage.Client
}

var FilterMap = map[string]StorageFilter{
	"images": StorageFilterImages,
	"all":    StorageFilterAll,
}

func conditionalMkdir(path string) error {
	// split dirs
	if i := strings.LastIndex(path, "/"); i <= 0 {
		return nil
	} else if err := os.MkdirAll(path[:i], os.FileMode(int(0777))); err != nil {
		return err
	}
	return nil
}

// client-side filter function
type StorageFilter func(attrs *storage.ObjectAttrs) bool

var (
	imageContentTypes = [3]string{"image/jpeg", "image/png", "image/svg+xml"}
)

// filter only image types
var StorageFilterImages = func(attrs *storage.ObjectAttrs) bool {
	for _, t := range imageContentTypes {
		if attrs.ContentType == t {
			return true
		}
	}
	return false
}

var StorageFilterAll = func(attrs *storage.ObjectAttrs) bool {
	return true
}

func NewAuthorizedStorageClient(ctx context.Context, aClient *AuthorizedHTTPClient) (storageClient *AuthorizedStorageClient, err error) {
	storageClient = &AuthorizedStorageClient{}
	storageClient.sClient, err = storage.NewClient(ctx, option.WithAuthCredentials(aClient.authCredentials),
		option.WithScopes(storage.ScopeReadOnly),
		storage.WithJSONReads())
	if err != nil {
		return nil, err
	}
	return storageClient, nil
}

// download from GCS storage bucket
func (aClient *AuthorizedStorageClient) StorageDownload(ctx context.Context, bucket string, prefix string, target string, filter StorageFilter) error {
	if err := os.MkdirAll(target, 0750); err != nil {
		return err
	}
	root, err := os.OpenRoot(target)
	if err != nil {
		return err
	}
	defer root.Close()

	work, ctx := errgroup.WithContext(ctx)
	var throttle = throttle.NewThrottle(4 * runtime.GOMAXPROCS(0))
	q := storage.Query{Prefix: prefix}
	if err := q.SetAttrSelection([]string{"Name", "ContentType"}); err != nil {
		panic(err)
	}
	bucketHandle := aClient.sClient.Bucket(bucket)
	it := bucketHandle.Objects(ctx, &q)
	for attrs, err := it.Next(); err != iterator.Done; attrs, err = it.Next() {
		if err != nil {
			panic(err)
		}
		work.Go(func() error {
			defer throttle.Done()
			throttle.Wait()
			objHandle := bucketHandle.Object(attrs.Name)
			outputFileName := attrs.Name
			if !filter(attrs) {
				return nil
			} else if err := rootMkdirAll(root, outputFileName); err != nil {
				return err
			} else if outF, err := root.Create(outputFileName); err != nil {
				return err
			} else if objReader, err := objHandle.NewReader(ctx); err != nil {
				return err
			} else if _, err := io.Copy(outF, objReader); err != nil {
				return err
			} else {
				err1 := objReader.Close()
				err2 := outF.Close()
				if err1 != nil {
					return err1
				}
				if err2 != nil {
					return err2
				}
				log.Printf("downloaded: %s\n", attrs.Name)
				return nil
			}
		})
	}
	if err := work.Wait(); err != nil {
		return err
	}
	return nil
}

func rootMkdirAll(root *os.Root, path string) error {
	dir := ppath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}
	parts := strings.Split(dir, "/")
	var current string
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = ppath.Join(current, part)
		err := root.Mkdir(current, 0750)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func (aClient *AuthorizedStorageClient) StorageUploadDirectory(ctx context.Context, bucketName, prefix, srcDir string) error {
	root, err := os.OpenRoot(srcDir)
	if err != nil {
		return err
	}
	defer root.Close()

	work, ctx := errgroup.WithContext(ctx)
	var throttle = throttle.NewThrottle(4 * runtime.GOMAXPROCS(0))
	err = filepath.WalkDir(srcDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // Skip directories
		}
		// Compute the object name in GCS
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		objectName := filepath.ToSlash(filepath.Join(prefix, relPath))
		// Upload to GCS
		work.Go(func() error {
			// Open the file
			defer throttle.Done()
			throttle.Wait()
			f, err := root.Open(relPath)
			if err != nil {
				return err
			}
			defer f.Close()
			wc := aClient.sClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
			if _, err := io.Copy(wc, f); err != nil {
				_ = wc.Close()
				return err
			}
			if err := wc.Close(); err != nil {
				return err
			}
			log.Printf("Uploaded %s to gs://%s/%s\n", path, bucketName, objectName)
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}
	if err := work.Wait(); err != nil {
		return err
	}
	return nil
}
