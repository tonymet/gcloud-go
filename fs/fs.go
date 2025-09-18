package fs

import (
	"context"
	"io/fs"
	"os"
	ppath "path"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	fbcompress "github.com/tonymet/gcloud-go/compress"
	"github.com/tonymet/gcloud-go/throttle"
)

type ShaRec struct {
	RelPath, Shasum string
	err             error
}

type ShaList []ShaRec

func ShaFiles(ctx context.Context, dirname, tempDir string) <-chan ShaRec {
	work, _ := errgroup.WithContext(ctx)
	throttle := throttle.NewThrottle(2 * runtime.GOMAXPROCS(0))
	shaChannel := make(chan ShaRec)
	go func() {
		defer close(shaChannel)
		err := filepath.WalkDir(dirname, func(path string, f fs.DirEntry, err error) error {
			if f.IsDir() {
				return nil
			}
			outFile := ppath.Join(tempDir, strings.ReplaceAll(path, "/", "__"))
			work.Go(func() error {
				defer throttle.Done()
				throttle.Wait()
				if h, err := fbcompress.HashAndCompressFile(outFile, path); err != nil {
					return err
				} else {
					s := ShaRec{ppath.Join("/", path), fbcompress.TextSum(h), nil}
					if err := os.Rename(outFile, ppath.Join(tempDir, s.Shasum)); err != nil {
						return err
					}
					shaChannel <- s
				}
				return nil
			})
			return nil
		})
		if err != nil {
			panic(err)
		}
		if err := work.Wait(); err != nil {
			panic(err)
		}
	}()
	return shaChannel
}
