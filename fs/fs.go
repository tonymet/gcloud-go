package fs

import (
	"io/fs"
	fbcompress "main/compress"
	"os"
	ppath "path"
	"path/filepath"
)

type ShaRec struct {
	RelPath, Shasum string
}

type ShaList []ShaRec

func ShaFiles(dirname, tempDir string) (ShaList, error) {
	var curShaList = make(ShaList, 0, 10)
	shaProcess := func(path string, f fs.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if h, err := fbcompress.CompressAndHashFile(path, ppath.Join(tempDir, f.Name())); err != nil {
			panic(err)
		} else {
			s := ShaRec{ppath.Join("/", path), fbcompress.TextSum(h)}
			if err := os.Rename(ppath.Join(tempDir, f.Name()), ppath.Join(tempDir, s.Shasum)); err != nil {
				panic(err)
			}
			curShaList = append(curShaList, s)
		}
		return nil
	} // walk files and update
	filepath.Walk(dirname, shaProcess)
	return curShaList, nil
}
