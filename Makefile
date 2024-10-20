@PHONY:bin
bin: build/bin/gcloud-go
build/bin/gcloud-go:main.go
	mkdir -p build/bin
	go build -o build/bin/gcloud-go -ldflags="-w -extldflags=-static" .

build/gcloud-go.tgz:build/bin/gcloud-go
	tar -zcf build/gcloud-go.tgz -C build/bin .

clean:
	rm -rf build/*
