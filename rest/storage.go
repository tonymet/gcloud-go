package rest

import (
	"context"
	"io"
	"os"
	ppath "path"
	"strings"
	"sync"

	"log"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
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

type StorageFilter func(attrs *storage.ObjectAttrs) bool

var StorageFilterImages = func(attrs *storage.ObjectAttrs) bool {
	return attrs.ContentType == "image/jpeg" || attrs.ContentType == "image/png"
}

func StorageDownload(creds *google.Credentials, bucket string, prefix string, target string, filter StorageFilter) error {
	var wgResults, wgWorkers sync.WaitGroup

	type objBundle struct {
		attrs  *storage.ObjectAttrs
		handle *storage.ObjectHandle
	}
	ctx := context.Background()
	if client, err := storage.NewClient(ctx, option.WithCredentials(creds)); err != nil {
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
		q.SetAttrSelection([]string{"Name", "Content-Type"})
		it := client.Bucket(bucket).Objects(ctx, &q)
		for attrs, err := it.Next(); err != iterator.Done; attrs, err = it.Next() {
			if err != nil {
				panic(err)
			}
			objHandle := client.Bucket("tonym.us").Object(attrs.Name)
			jobs <- objBundle{attrs, objHandle}
		}
		close(jobs)
		wgWorkers.Wait()
		close(results)
		wgResults.Wait()
	}
	return nil
}