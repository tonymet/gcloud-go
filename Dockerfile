FROM golang:alpine AS build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o build/bin/gcloud-go -ldflags="-w -extldflags=-static" .
FROM scratch
# Copy binary from build step
VOLUME /tmp
VOLUME /content
COPY --from=build /workspace/build/bin/gcloud-go ./
ENTRYPOINT ["/gcloud-go"]
