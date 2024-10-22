package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var countTest int

func myWalk(path string, f fs.FileInfo, err error) error {
	countTest++
	return nil
}
func countFiles(_ os.DirEntry) {
	countTest++
}
func Test_processDir(t *testing.T) {
	filepath.Walk("../../public", myWalk)
	expect := 10
	if countTest != expect {
		t.Logf("expect: 6, got %d\n", countTest)
		t.Fail()

	}
}

func Test_shaFiles(t *testing.T) {
	type args struct {
		dirname, tempName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"test1",
			args{"../../public", "temp"},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if s, err := ShaFiles(&sync.WaitGroup{}, tt.args.dirname, tt.args.tempName); (err != nil) != tt.wantErr {
				t.Errorf("shaFiles() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("%+v\n", s)
			}
		})
	}
}
