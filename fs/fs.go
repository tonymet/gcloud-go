package fs

import (
	"context"
	"io/fs"
	"os"
	ppath "path"
	"path/filepath"
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

func ShaFiles(dirname, tempDir string) <-chan ShaRec {
	work, _ := errgroup.WithContext(context.Background())
	throttle := throttle.NewThrottle(16)
	shaChannel := make(chan ShaRec)
	go func() {
		err := filepath.WalkDir(dirname, func(path string, f fs.DirEntry, err error) error {
			if f.IsDir() {
				return nil
			}
			outFile := ppath.Join(tempDir, strings.Replace(path, "/", "__", -1))
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
		close(shaChannel)
	}()
	return shaChannel
}
