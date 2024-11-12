package rest

import (
	"cloud.google.com/go/storage"
	"testing"
)

type testParams struct {
	arg  storage.ObjectAttrs
	want bool
}

var testFixtures = []testParams{
	{storage.ObjectAttrs{ContentType: "image/png"}, true},
	{storage.ObjectAttrs{ContentType: "application/binary"}, false},
	{storage.ObjectAttrs{ContentType: "image/svg+xml"}, true},
}

func BenchmarkStorageFilterImagesRegex(b *testing.B) {
	for range b.N {
		storageFilterImagesRegex(&testFixtures[0].arg)
	}
}
func BenchmarkStorageFilterImages(b *testing.B) {
	for range b.N {
		StorageFilterImages(&testFixtures[0].arg)
	}
}
func TestStorageFilterImages(t *testing.T) {
	for _, f := range testFixtures {
		if !(StorageFilterImages(&f.arg) == f.want && storageFilterImagesRegex(&f.arg) == f.want) {
			t.Logf("expect: %t, arg: %s \n", f.want, f.arg.ContentType)
			t.Fail()
		}
	}
}
