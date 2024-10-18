
## About

gcloud-go -- a fast & lightweight rest client for deployment to Firebase Hosting

For VM images, containers & cloud functions that have demanding iops, vcpu or
time requirements, gcloud-go will avoid unnecessary resources

### Features
* concurrent & resume-able uploads. Only modified files are uploaded during release
* uses native authentication / metadata service within the cloud
* docker image and linux-amd64 binaries available (see below)

### In Progress
* \> 1000 file / paginated upload support
* multi-core file signing & compression
* gcs storage & firebase storage support for large files
* google cloud pub/sub support for triggering builds


## Compared to Firebase-tools

| docker image  | size   | savings  |  
|---|---|---|
| firebase-tools  | 245mb  | n/a   |   
|  gcloud-go | 19mb  |  92%  |   


## Authenticating
1. Be sure to have a service account and the Firebase Hosting API enabled in your GCP Project.  [See the docs](https://firebase.google.com/docs/hosting/api-deploy)
2. Within GCP, the Metadata service will be used. Make sure your VM , Container, etc has the appropriate service account attached with privileges for Firebase hosting API (see above)
3. Outside GCP, you will need a service account key . $GOOGLE_APPLICATION_CREDENTIALS env var should reference that key.  See the docs above for generating a service account key.


## Installation
To install inside your terminal, download and copy to your path.
```
# download
$ curl -O https://github.com/tonymet/gcloud-go/releases/download/2024-10-18-b096696/gcloud-go-2024-10-18-b096696-linux-x86_64.tar.gz
# un-tar
$ tar -zxf *tar.gz
# run the binary
$ ./gcloud-go
```


## Docker Images
### Example Run
```
	$ docker run -v$HOME/public:/content -v$(pwd):/src \
    us-west1-docker.pkg.dev/tonym-us/gcloud-lite/gcloud-go \
    -source /content -cred /src/service-ident-3xxxc42.json; 
```

### Example Build
For use inside docker images, copy the pre-built binary using the command `COPY --image=`
You can run the command inside your scripts by calling `/gcloud-go`

```
FROM alpine AS app-env
RUN apk update && apk add --no-cache zsh bind-tools envsubst tzdata
COPY --from=us-west1-docker.pkg.dev/tonym-us/gcloud-lite/gcloud-go /gcloud-go /gcloud-go
COPY . /app
WORKDIR /app
CMD ["zsh", "build.sh"]
```


## Usage
#### Example
```
$ gcloud-go deploy -site dev-isgithubipv6 -source content
```
#### Full usage
```
usage: ./gcloud-go deploy [options]
 options:
  -connections int
        number of connections (default 8)
  -cred string
        path to service principal. Use ENV var GOOGLE_APPLICATION_CREDENTAILS by default. Within GCP, metadata server will be used
  -site string
        Name of site (not project) (default "default")
  -source string
        Source directory for content (default "content")
  -temp string
        temp directory for staging files prior to upload (default "/tmp")

```


## Developer Docs
### Building
```
$ make bin
mkdir -p build/bin
go build -o build/bin/gcloud-go -ldflags="-extldflags=-static" .
```

## Related Projects
* [gcloud-lite](https://github.com/tonymet/gcloud-lite) -- Stripped gcloud cli for resource-constrained environments