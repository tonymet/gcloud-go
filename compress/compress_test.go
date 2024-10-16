package compress

import (
	"fmt"
	"hash"
	"testing"
)

func textSum(h *hash.Hash) string {
	return fmt.Sprintf("%x", (*h).Sum(nil))
}

/*

func Test_compressFile(t *testing.T) {
	type args struct {
		inFile  string
		outFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"test1",
			args{"test1.txt", "test1.txt.gz"},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := compressFile(tt.args.inFile, tt.args.outFile); (err != nil) != tt.wantErr {
				t.Errorf("compressFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/

func Test_compressAndHashFile(t *testing.T) {
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
			"c9178746617152f284e64349f749dcfe4b3805041283153f0c74fc5a8559b204",
		},
		// TODO: Add test cases.
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compressAndHashFile(tt.args.inFile, tt.args.outFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("compressAndHashFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if textSum(got) != tt.sum {
				t.Errorf("compressAndHashFile() = %v, want %v", textSum(got), tt.sum)
			}
		})
	}
}
