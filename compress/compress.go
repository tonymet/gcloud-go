package compress

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
)

func TextSum(h *hash.Hash) string {
	return fmt.Sprintf("%x", (*h).Sum(nil))
}

func hashAndCompressCopy(target io.WriteCloser, source io.Reader) (*hash.Hash, error) {
	h := sha256.New()
	mWriter := io.MultiWriter(target, h)
	zWriter := gzip.NewWriter(mWriter)
	defer zWriter.Close() //nolint:errcheck
	if _, err := io.Copy(zWriter, source); err != nil {
		return nil, err
	} else {
		return &h, nil
	}
}

// compress file
func HashAndCompressFile(outFile, inFile string) (*hash.Hash, error) {
	var (
		inF  fileReader
		outF fileWriter
		err  error
	)
	if inF, err = fs.Open(inFile); err != nil {
		panic(err)
	} else if outF, err = fs.Create(outFile); err != nil {
		panic(err)
	}
	defer inF.Close()  //nolint:errcheck
	defer outF.Close() //nolint:errcheck
	if h, err := hashAndCompressCopy(outF, inF); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}
