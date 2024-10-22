package compress

import (
	"io"
	"strings"
	"testing"
)

func Test_compressAndHashCopy(t *testing.T) {
	type args struct {
		target io.WriteCloser
		source io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		sum     string
	}{
		{
			"test1",
			args{discardWriter, strings.NewReader("asdfasdf\n")},
			false,
			"29570ad7ec76d864317b8fe582d43bab1493fe445be76c6cb2b024ffb0fb5625",
		},
		// TODO: Add test cases.
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashCompressCopy(tt.args.target, tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("compressAndHashFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if TextSum(got) != tt.sum {
				t.Errorf("compressAndHashFile() = %v, want %v", TextSum(got), tt.sum)
			}
		})
	}
}

func Test_compressAndHashFile(t *testing.T) {
	t.Skip("manual testing")
	type args struct {
		inFile  string
		outFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		sum     string
	}{
		{
			"test1",
			args{"test1.txt", "test1.txt.gz"},
			false,
			// "asdfasdf\n"
			"29570ad7ec76d864317b8fe582d43bab1493fe445be76c6cb2b024ffb0fb5625",
		},
		// TODO: Add test cases.
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashAndCompressFile(tt.args.outFile, tt.args.inFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("compressAndHashFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if TextSum(got) != tt.sum {
				t.Errorf("compressAndHashFile() = %v, want %v", TextSum(got), tt.sum)
			}
		})
	}
}
