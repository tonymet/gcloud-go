package fs

import (
	"io/fs"
	fbcompress "main/compress"
	"os"
	"path/filepath"
)

type ShaRec struct {
	RelPath, Shasum string
}

type ShaList []ShaRec

func ShaFiles(dirname string) (ShaList, error) {
	var curShaList = make(ShaList, 0, 10)
	shaProcess := func(path string, f fs.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if h, err := fbcompress.CompressAndHashFile(path, "temp/"+f.Name()); err != nil {
			panic(err)
		} else {
			s := ShaRec{f.Name(), fbcompress.TextSum(h)}
			if err := os.Rename("temp/"+f.Name(), "temp/"+s.Shasum); err != nil {
				panic(err)
			}
			curShaList = append(curShaList, s)
		}
		return nil
	} // walk files and update
	filepath.Walk(dirname, shaProcess)
	return curShaList, nil
}
