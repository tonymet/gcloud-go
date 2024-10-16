package compress

import "testing"

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
