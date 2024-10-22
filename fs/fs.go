package fs

import (
	"io/fs"
	fbcompress "main/compress"
	"os"
	ppath "path"
	"path/filepath"
	"sync"
)

type ShaRec struct {
	RelPath, Shasum string
	err             error
}

type ShaList []ShaRec

func ShaFiles(wg *sync.WaitGroup, dirname, tempDir string) (<-chan ShaRec, error) {
	shaChannel := make(chan ShaRec)
	shaProcess := func(path string, f fs.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if h, err := fbcompress.HashAndCompressFile(ppath.Join(tempDir, f.Name()), path); err != nil {
			panic(err)
		} else {
			s := ShaRec{ppath.Join("/", path), fbcompress.TextSum(h), nil}
			if err := os.Rename(ppath.Join(tempDir, f.Name()), ppath.Join(tempDir, s.Shasum)); err != nil {
				panic(err)
			}
			shaChannel <- s
		}
		return nil
	} // walk files and update
	wg.Add(1)
	go func() {
		defer wg.Done()
		filepath.Walk(dirname, shaProcess)
		close(shaChannel)
	}()
	return shaChannel, nil
}
