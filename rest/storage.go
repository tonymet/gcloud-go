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
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

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

// download from GCS storage bucket
func (aClient *AuthorizedHTTPClient) StorageDownload(bucket string, prefix string, target string, filter StorageFilter) error {
	var wgResults, wgWorkers sync.WaitGroup

	type objBundle struct {
		attrs  *storage.ObjectAttrs
		handle *storage.ObjectHandle
	}
	ctx := context.Background()
	if sClient, err := storage.NewClient(ctx, option.WithAuthCredentials(aClient.authCredentials),
		option.WithScopes(storage.ScopeReadOnly),
		storage.WithJSONReads()); err != nil {
		return err
	} else {
		imageWorker := func(jobs <-chan objBundle, results chan<- error) {
			defer wgWorkers.Done()
			for j := range jobs {
				outputFileName := ppath.Join(target, j.attrs.Name)
				if !filter(j.attrs) {
					results <- nil
				} else if err := conditionalMkdir(outputFileName); err != nil {
					results <- err
				} else if outF, err := os.Create(outputFileName); err != nil {
					results <- err
				} else if objReader, err := j.handle.NewReader(ctx); err != nil {
					results <- err
				} else if _, err := io.Copy(outF, objReader); err != nil {
					results <- err
				} else {
					objReader.Close()
					outF.Close()
					results <- nil
					log.Printf("downloaded: %s\n", j.attrs.Name)
				}
			}
		}
		jobs, results := make(chan objBundle), make(chan error)
		for w := 1; w <= 8; w++ {
			wgWorkers.Add(1)
			go imageWorker(jobs, results)
		}
		wgResults.Add(1)
		go func() {
			defer wgResults.Done()
			for {
				if res, ok := <-results; !ok {
					break
				} else if res != nil {
					panic(res)
				}
			}
		}()
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
			objHandle := bucketHandle.Object(attrs.Name)
			jobs <- objBundle{attrs, objHandle}
		}
		close(jobs)
		wgWorkers.Wait()
		close(results)
		wgResults.Wait()
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
		return filepath.WalkDir(srcDir, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil // Skip directories
			}
			// Compute the object name in GCS
			objectName := filepath.ToSlash(filepath.Join(prefix, path))
			// Open the file
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			// Upload to GCS
			wc := sClient.Bucket(bucketName).Object(objectName).NewWriter(ctx)
			if _, err := io.Copy(wc, f); err != nil {
				wc.Close()
				return err
			}
			if err := wc.Close(); err != nil {
				return err
			}
			log.Printf("Uploaded %s to gs://%s/%s\n", path, bucketName, objectName)
			return nil
		})
	}
}
