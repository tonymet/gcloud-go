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
	type jobArgs struct {
		path string
		f    fs.FileInfo
		err  error
	}
	var curShaList = make(ShaList, 0, 10)
	shaProcess := func(jobs <-chan jobArgs, results chan<- ShaRec) error {
		for j := range jobs {
			path, f, _ := j.path, j.f, j.err
			if f.IsDir() {
				continue
			}
			if h, err := fbcompress.CompressAndHashFile(path, ppath.Join(tempDir, f.Name())); err != nil {
				panic(err)
			} else {
				s := ShaRec{ppath.Join("/", path), fbcompress.TextSum(h)}
				if err := os.Rename(ppath.Join(tempDir, f.Name()), ppath.Join(tempDir, s.Shasum)); err != nil {
					panic(err)
				}
				results <- s
			}
		}
		return nil
	} // walk files and update

	// set up
	jobs, results := make(chan jobArgs), make(chan ShaRec)

	queueFile := func(path string, f fs.FileInfo, err error) error {
		jobs <- jobArgs{path, f, err}
		return nil
	}

	for w := 1; w <= 4; w++ {
		go shaProcess(jobs, results)
	}

	go func() {
		filepath.Walk(dirname, queueFile)
		close(jobs)
	}()

	for {
		if s, ok := <-results; !ok {
			break
		} else {
			curShaList = append(curShaList, s)
		}
	}
	// close jobs
	return curShaList, nil
}
