
## About

gcloud-go -- a fast & lightweight rest client for deployment to Firebase Hosting, Google Cloud Storage and other APIs

For VM images, containers & cloud functions that have demanding iops, vcpu or
time requirements, gcloud-go will avoid unnecessary resources

### Features
* concurrent & incremental uploads. Only modified files are uploaded during release
* uses native authentication / metadata service within the cloud
* multi-core file hashing & compression
* docker image and linux-amd64 binaries available (see below)
* sync 10k files in seconds with fixed memory
* GCS storage downloads

### In Progress
* google cloud pub/sub support for triggering builds

## Compared to Firebase-tools

#### Size Comparison 

| docker image  | size   | savings  |  
|---|---|---|
| firebase-tools  | 245mb  | n/a   |   
|  gcloud-go | 19mb  |  92%  |   

#### Speed Comparison

| docker image  | deploy time for 7500 files   | savings  |  
|---|---|---|
| firebase-tools  | 16.5s  | n/a   |   
|  gcloud-go | 8.8s |  45%  |   


## Authenticating
1. Be sure to have a service account and the Firebase Hosting API enabled in your GCP Project.  [See the docs](https://firebase.google.com/docs/hosting/api-deploy)
2. Within GCP, the Metadata service will be used. Make sure your VM , Container, etc has the appropriate service account attached with privileges for Firebase hosting API (see above)
3. Outside GCP, use a service account & impersonation to run the utility as a service account while logged in as a priviledged user. See *Impersonating A Service Account* . 

[Google Docs on Application Default
Credentials](https://cloud.google.com/docs/authentication/application-default-credentials)
explains the environment variables & configuration files that are used to find
credentials


## Installation

### Install From Binary
See [releases page for the latest builds](https://github.com/tonymet/gcloud-go/releases)

To install inside your terminal, download and copy to your path.
```
# download
$ curl -LO https://github.com/tonymet/gcloud-go/releases/download/2024-10-18-d5dc06a/gcloud-go-2024-10-18-d5dc06a-linux-x86_64.tar.gz
# un-tar
$ tar -zxf *tar.gz
# run the binary
$ ./gcloud-go
```

### Install From Source
```
$ go install github.com/tonymet/gcloud-go/cmd/gcloud-go
```


## Docker Images
See [Google Artifact Registry](https://us-west1-docker.pkg.dev/tonym-us/gcloud-lite/gcloud-go) (similar to Docker Hub) for docker images.

### Example Run
```
# run inside GCE using metadata-based credentials
$ docker run -v $(pwd)/test-output-small:/content \
    gcloud-go deploy -source /content -site $SITE_NAME


# run outside GCE with APPL
$ docker run -eGOOGLE_APPLICATION_CREDENTIALS=$GOOGLE_APPLICATION_CREDENTIALS \
      -v$HOME/public:/content -v$(pwd):/src \
      us-west1-docker.pkg.dev/tonym-us/gcloud-lite/gcloud-go \
      deploy -source /content -site $SITE_NAME;
```

### Example Build
For use inside docker images, copy the pre-built binary using the command `COPY --from=`
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
usage: ./gcloud-go COMMAND [options]
 deploy:
  -connections int
        number of connections (default 8)
  -site string
        Name of site (not project) (default "default")
  -source string
        Source directory for content (default "content")
  -temp string
        temp directory for staging files prior to upload (default "/tmp")

storage:
  -bucket string
        GCS Bucket
  -prefix string
        GCS Object Prefix (default "/")
  -target string
        Target Directory for download (default ".")

```


## Developer Docs
### Building
```
$ make bin
mkdir -p build/bin
go build -o build/bin/gcloud-go -ldflags="-extldflags=-static" .
```

### Impersonating a Service Account
It's preferred to always run the utility as a service account.  To avoid risky service key creation, you can use *impersonation*
This will log in as a priviledged user, but authorize the utility as a service account. The temporary credential will
be written to `$HOME/.config/gcloud/application_default_credentials.json `
e.g.

```
PROJECT=your-project-1234
gcloud auth application-default login --impersonate-service-account  dev-gcloud-go@${PROJECT}.iam.gserviceaccount.com     
```

## Related Projects
* [gcloud-lite](https://github.com/tonymet/gcloud-lite) -- Stripped gcloud cli for resource-constrained environments
