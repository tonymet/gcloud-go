FROM golang:alpine AS build
WORKDIR /app/src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/bin/firebase -ldflags="-extldflags=-static" .
FROM scratch
# Copy binary from build step
COPY --from=build /app/bin/firebase ./