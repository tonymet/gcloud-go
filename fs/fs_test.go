package fs

import (
	"io/fs"
	"path/filepath"
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
