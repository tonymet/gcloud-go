static: firebase 
	go build -o firebase -ldflags="-extldflags=-static" .

firebase:main.go

docker_run:
	docker run -v/home/tonymet/sotion/isgithubipv6.lol/public:/content -v$(pwd):/src firebase-go -source /content -cred /src/tonym-us-311af670bc42.json;
