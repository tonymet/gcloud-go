static: firebase 
	go build -o firebase -ldflags="-extldflags=-static" .

firebase:oauth.go