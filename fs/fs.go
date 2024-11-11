package fs

import (
	"io/fs"
	fbcompress "main/compress"
	"os"
	ppath "path"
	"path/filepath"
	"strings"
	"sync"
)

type ShaRec struct {
	RelPath, Shasum string
	err             error
}

type ShaList []ShaRec

func ShaFiles(dirname, tempDir string) (<-chan ShaRec, func(), error) {
	type jobRequest struct {
		outFile, inFile string
	}
	var wgJobs sync.WaitGroup
	shaChannel := make(chan ShaRec)
	jobsChannel := make(chan jobRequest)
	worker := func() {
		defer wgJobs.Done()
		for j := range jobsChannel {
			if h, err := fbcompress.HashAndCompressFile(j.outFile, j.inFile); err != nil {
				panic(err)
			} else {
				s := ShaRec{ppath.Join("/", j.inFile), fbcompress.TextSum(h), nil}
				if err := os.Rename(j.outFile, ppath.Join(tempDir, s.Shasum)); err != nil {
					panic(err)
				}
				shaChannel <- s
			}
		}
	}
	wgJobs.Add(4)
	for i := 0; i < 4; i++ {
		go worker()
	}
	shaProcess := func(path string, f fs.DirEntry, err error) error {
		if f.IsDir() {
			return nil
		}
		jobsChannel <- jobRequest{ppath.Join(tempDir, strings.Replace(path, "/", "__", -1)), path}
		return nil
	} // walk files and update
	execFunc := func() {
		if err := filepath.WalkDir(dirname, shaProcess); err != nil {
			panic(err)
		}
		close(jobsChannel)
		wgJobs.Wait()
		close(shaChannel)
	}
	return shaChannel, execFunc, nil
}
