package compress

import (
	"io"
	"os"
	"strings"
)

// discard WriterCloser for testing
type discardCloser struct {
	io.Writer
}

func (discardCloser) Close() error {
	return nil
}

var discardWriter = discardCloser{io.Discard}

// generic fileReader & fileWriter to avoid build-time dep on "os"
type fileReader interface {
	io.Closer
	io.Reader
	io.ReaderAt
}
type fileWriter interface {
	io.WriteCloser
	//io.ReaderAt
}

// strings.Reader --> ReaderCloser
type stringReaderCloser struct {
	*strings.Reader
}

func (stringReaderCloser) Close() error {
	return nil
}

func NewReaderCloser(contents string) fileReader {
	return stringReaderCloser{strings.NewReader(contents)}
}

// for run-time to call os methods
type osFS struct{}

// default for run-time
var fs fileSystem = osFS{}

// mock FS methods interface
type fileSystem interface {
	Open(name string) (fileReader, error)
	Create(name string) (fileWriter, error)
}

func (osFS) Open(name string) (fileReader, error)   { return os.Open(name) }
func (osFS) Create(name string) (fileWriter, error) { return os.Create(name) }

// for test-time replace os.File contents with string
type stringFS struct {
	contents string
}

func (sfs stringFS) Open(name string) (fileReader, error) {
	return fileReader(NewReaderCloser(sfs.contents)), nil
}
func (stringFS) Create(name string) (fileWriter, error) { return discardWriter, nil }
