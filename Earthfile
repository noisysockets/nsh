VERSION 0.8
FROM golang:1.22-bookworm
WORKDIR /workspace

all:
  ARG VERSION=dev
  BUILD --platform=linux/amd64 --platform=linux/arm64 +docker
  COPY (+build/nsh --GOARCH=amd64) ./dist/nsh-linux-amd64
  COPY (+build/nsh --GOARCH=arm64) ./dist/nsh-linux-arm64
  COPY (+build/nsh --GOOS=darwin --GOARCH=amd64) ./dist/nsh-darwin-amd64
  COPY (+build/nsh --GOOS=darwin --GOARCH=arm64) ./dist/nsh-darwin-arm64
  COPY (+build/nsh --GOOS=windows --GOARCH=amd64) ./dist/nsh-windows-amd64.exe
  RUN cd dist && find . -type f -exec sha256sum {} \; >> ../checksums.txt
  SAVE ARTIFACT ./dist/nsh-linux-amd64 AS LOCAL dist/nsh-linux-amd64
  SAVE ARTIFACT ./dist/nsh-linux-arm64 AS LOCAL dist/nsh-linux-arm64
  SAVE ARTIFACT ./dist/nsh-darwin-amd64 AS LOCAL dist/nsh-darwin-amd64
  SAVE ARTIFACT ./dist/nsh-darwin-arm64 AS LOCAL dist/nsh-darwin-arm64
  SAVE ARTIFACT ./dist/nsh-windows-amd64.exe AS LOCAL dist/nsh-windows-amd64.exe
  SAVE ARTIFACT ./checksums.txt AS LOCAL dist/checksums.txt

docker:
  FROM debian:bookworm-slim
  RUN groupadd -g 65532 nonroot \
    && useradd -u 65532 -g 65532 -s /sbin/nologin -m nonroot
  # We need a ping SUID binary for ICMP forwarding to work.
  RUN apt update \
      && apt install -y iputils-ping \
      && rm -rf /var/lib/apt/lists/*
  COPY LICENSE /usr/local/share/nsh/
  ARG TARGETARCH
  ENV container=docker
  COPY (+build/nsh --GOOS=linux --GOARCH=${TARGETARCH}) /nsh
  USER 65532:65532
  ENTRYPOINT ["/nsh"]
  ARG VERSION=dev
  SAVE IMAGE --push ghcr.io/noisysockets/nsh:${VERSION}
  SAVE IMAGE --push ghcr.io/noisysockets/nsh:latest

build:
  ARG GOOS=linux
  ARG GOARCH=amd64
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  ARG VERSION=dev
  RUN --secret TELEMETRY_TOKEN=telemetry_token \
    CGO_ENABLED=0 go build -o nsh --ldflags "-s \
    -X 'github.com/noisysockets/nsh/internal/constants.Version=${VERSION}' \
    -X 'github.com/noisysockets/nsh/internal/constants.TelemetryToken=${TELEMETRY_TOKEN}'"
  SAVE ARTIFACT ./nsh AS LOCAL dist/nsh-${GOOS}-${GOARCH}

tidy:
  LOCALLY
  RUN go mod tidy
  RUN go fmt ./...
  RUN for dir in $(find . -name 'go.mod'); do \
      (cd "${dir%/go.mod}" && go mod tidy); \
    done

lint:
  FROM golangci/golangci-lint:v1.57.2
  WORKDIR /workspace
  COPY . ./
  RUN golangci-lint run --timeout 5m ./...

test:
  COPY . ./
  RUN go test -coverprofile=coverage.out -v ./...
  SAVE ARTIFACT ./coverage.out AS LOCAL coverage.out