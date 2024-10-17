FROM golang:alpine AS build
RUN apk --no-cache add ca-certificates
WORKDIR /app/src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/bin/firebase -ldflags="-extldflags=-static" .
RUN touch /tmp/.tmp
FROM scratch
# Copy binary from build step
COPY --from=build /tmp/.tmp /tmp/.tmp
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/bin/firebase ./
ENTRYPOINT ["/firebase"]