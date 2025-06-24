// GCS Google Cloud Storage API
package rest

import (
	"context"
	"io"
	"io/fs"
	"os"
	ppath "path"
	"path/filepath"
	"strings"
	"sync"

	"log"

	"cloud.google.com/go/storage"
	"github.com/tonymet/gcloud-go/throttle"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

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

// download from GCS storage bucket
func (aClient *AuthorizedHTTPClient) StorageDownload(bucket string, prefix string, target string, filter StorageFilter) error {
	var wgWorkers sync.WaitGroup
	var throttle = throttle.NewThrottle(8)
	ctx := context.Background()
	if sClient, err := storage.NewClient(ctx, option.WithAuthCredentials(aClient.authCredentials),
		option.WithScopes(storage.ScopeReadOnly),
		storage.WithJSONReads()); err != nil {
		return err
	} else {
		q := storage.Query{Prefix: prefix}
		if err := q.SetAttrSelection([]string{"Name", "ContentType"}); err != nil {
			panic(err)
		}
		bucketHandle := sClient.Bucket(bucket)
		it := bucketHandle.Objects(ctx, &q)
		for attrs, err := it.Next(); err != iterator.Done; attrs, err = it.Next() {
			if err != nil {
				panic(err)
			}
			wgWorkers.Add(1)
			go func() {
				defer wgWorkers.Done()
				defer throttle.Done()
				throttle.Wait()
				objHandle := bucketHandle.Object(attrs.Name)
				outputFileName := ppath.Join(target, attrs.Name)
				if !filter(attrs) {
					return
				} else if err := conditionalMkdir(outputFileName); err != nil {
					panic(err)
				} else if outF, err := os.Create(outputFileName); err != nil {
					panic(err)
				} else if objReader, err := objHandle.NewReader(ctx); err != nil {
					panic(err)
				} else if _, err := io.Copy(outF, objReader); err != nil {
					panic(err)
				} else {
					objReader.Close()
					outF.Close()
					log.Printf("downloaded: %s\n", attrs.Name)
				}
			}()
		}
		wgWorkers.Wait()
	}
	return nil
}

func (aClient *AuthorizedHTTPClient) StorageUploadDirectory(bucketName, prefix, srcDir string) error {
	ctx := context.Background()
	if sClient, err := storage.NewClient(ctx, option.WithAuthCredentials(aClient.authCredentials),
		option.WithScopes(storage.ScopeReadWrite),
		storage.WithJSONReads()); err != nil {
		return err
	} else {
		var wgWorker sync.WaitGroup
		var throttle = throttle.NewThrottle(8)
		err := filepath.WalkDir(srcDir, func(path string, info fs.DirEntry, err error) error {
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
			wgWorker.Add(1)
			go func() {
				// Open the file
				defer wgWorker.Done()
				defer throttle.Done()
				throttle.Wait()
				f, err := os.Open(path)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				wc := sClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
				if _, err := io.Copy(wc, f); err != nil {
					wc.Close()
					panic(err)
				}
				if err := wc.Close(); err != nil {
					panic(err)
				}
				log.Printf("Uploaded %s to gs://%s/%s\n", path, bucketName, objectName)
			}()
			return nil
		})
		wgWorker.Wait()
		return err
	}
}
