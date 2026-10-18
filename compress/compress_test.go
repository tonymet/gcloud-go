package compress

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"testing"
)

type bufferCloser struct {
	*bytes.Buffer
}

func (bufferCloser) Close() error {
	return nil
}

type memoryFS struct {
	files map[string]*bytes.Buffer
}

func newMemoryFS() *memoryFS {
	return &memoryFS{files: make(map[string]*bytes.Buffer)}
}

func (m *memoryFS) Open(name string) (fileReader, error) {
	buf, ok := m.files[name]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", name)
	}
	return stringReaderCloser{strings.NewReader(buf.String())}, nil
}

func (m *memoryFS) Create(name string) (fileWriter, error) {
	buf := &bytes.Buffer{}
	m.files[name] = buf
	return bufferCloser{buf}, nil
}

func decompressGzip(data []byte) (string, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer zr.Close()
	decompressed, err := io.ReadAll(zr)
	if err != nil {
		return "", err
	}
	return string(decompressed), nil
}

func Test_compressAndHashCopy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
	}{
		{
			name:    "test1",
			input:   "asdfasdf\n",
			wantErr: false,
		},
		{
			name:    "test2",
			input:   "xasdfasdf\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := hashAndCompressCopy(bufferCloser{&buf}, strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("hashAndCompressCopy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify the calculated checksum matches the SHA256 of the compressed output
			expectedHash := fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
			if TextSum(got) != expectedHash {
				t.Errorf("TextSum(got) = %v, want %v", TextSum(got), expectedHash)
			}

			// Verify round-trip decompression matches the original input
			decompressed, err := decompressGzip(buf.Bytes())
			if err != nil {
				t.Fatalf("failed to decompress gzip output: %v", err)
			}
			if decompressed != tt.input {
				t.Errorf("decompressed content = %q, want %q", decompressed, tt.input)
			}
		})
	}
}

func Test_compressAndHashFile(t *testing.T) {
	tests := []struct {
		name    string
		inFile  string
		outFile string
		content string
		wantErr bool
	}{
		{
			name:    "testGood",
			inFile:  "test1.txt",
			outFile: "test1.txt.gz",
			content: "asdfasdf\n",
			wantErr: false,
		},
		{
			name:    "testOther",
			inFile:  "test2.txt",
			outFile: "test2.txt.gz",
			content: "xasdfasdf\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := newMemoryFS()
			memFS.files[tt.inFile] = bytes.NewBufferString(tt.content)
			fs = memFS

			got, err := HashAndCompressFile(tt.outFile, tt.inFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashAndCompressFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			outBuf := memFS.files[tt.outFile]
			if outBuf == nil {
				t.Fatalf("expected output file %s was not written", tt.outFile)
			}

			// Verify checksum matches compressed data
			expectedHash := fmt.Sprintf("%x", sha256.Sum256(outBuf.Bytes()))
			if TextSum(got) != expectedHash {
				t.Errorf("TextSum(got) = %v, want %v", TextSum(got), expectedHash)
			}

			// Verify round-trip decompression
			decompressed, err := decompressGzip(outBuf.Bytes())
			if err != nil {
				t.Fatalf("failed to decompress gzip file: %v", err)
			}
			if decompressed != tt.content {
				t.Errorf("decompressed content = %q, want %q", decompressed, tt.content)
			}
		})
	}
}
