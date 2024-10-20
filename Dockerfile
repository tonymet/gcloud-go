FROM golang:alpine AS build
RUN apk --no-cache add ca-certificates
WORKDIR /app/src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/bin/gcloud-go -ldflags="-w -extldflags=-static" .
RUN touch /tmp/.tmp
FROM scratch
# Copy binary from build step
COPY --from=build /tmp/.tmp /tmp/.tmp
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/bin/gcloud-go ./
ENTRYPOINT ["/gcloud-go"]