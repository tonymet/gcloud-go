package compress

import (
	"compress/gzip"
	"crypto/sha256"
	"hash"
	"io"
	"os"
)

// compress file
func compressAndHashFile(inFile, outFile string) (*hash.Hash, error) {
	h := sha256.New()
	if inF, err := os.Open(inFile); err != nil {
		panic(err)
	} else {
		defer inF.Close()
		if outF, err := os.Create(outFile); err != nil {
			panic(err)
		} else {
			defer outF.Close()
			mWriter := io.MultiWriter(outF, h)
			zWriter := gzip.NewWriter(mWriter)
			defer zWriter.Close()
			// copy to gzip
			if _, err := io.Copy(zWriter, inF); err != nil {
				panic(err)
			}
		}
	}
	return &h, nil
}
