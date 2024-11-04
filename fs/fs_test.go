package fs

import (
	"io/fs"
	"path/filepath"
	"sync"
	"testing"
)

func Test_processDir(t *testing.T) {
	t.Skip("manual test")
	var countTest int

	myWalk := func(path string, f fs.FileInfo, err error) error {
		countTest++
		return nil
	}

	if err := filepath.Walk("../test-output-small", myWalk); err != nil {
		panic(err)
	}
	if countTest != 462 {
		t.Logf("expect: 462, got %d\n", countTest)
		t.Fail()

	}
}

func Test_shaFiles(t *testing.T) {
	t.Skip("manual test")
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
			args{"../test-output-small", "/tmp/kaka"},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			if s, err := ShaFiles(&wg, tt.args.dirname, tt.args.tempName); (err != nil) != tt.wantErr {
				sRec := <-s
				t.Errorf("shaFiles() error = %v, wantErr %v", sRec.err, tt.wantErr)
			} else {
				t.Logf("%+v\n", s)
			}
			wg.Wait()
		})
	}
}
