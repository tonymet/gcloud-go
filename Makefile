build/bin/gcloud-go:main.go
	mkdir -p build
	go build -o build/bin/gcloud-go -ldflags="-extldflags=-static" .

build/gcloud-go.tgz:build/bin/gcloud-go
	tar -zcf build/gcloud-go.tgz -C build/bin .

docker_run:
	docker run -v/home/tonymet/sotion/isgithubipv6.lol/public:/content -v$(pwd):/src firebase-go -source /content -cred /src/tonym-us-311af670bc42.json;
