package compress

import (
	"compress/gzip"
	"io"
	"os"
)

// compress file
func compressFile(inFile, outFile string) error {
	if inF, err := os.Open(inFile); err != nil {
		panic(err)
	} else {
		defer inF.Close()
		if outF, err := os.Create(outFile); err != nil {
			panic(err)
		} else {
			defer outF.Close()
			zWriter := gzip.NewWriter(outF)
			defer zWriter.Close()
			if _, err := io.Copy(zWriter, inF); err != nil {
				panic(err)
			}
		}
	}
	return nil
}
