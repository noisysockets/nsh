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
  FROM registry.dpeckett.dev/debian:bookworm-ultraslim
  ENV container=docker
  COPY LICENSE /usr/share/doc/nsh/copyright
  ARG TARGETARCH
  COPY (+build/nsh --GOOS=linux --GOARCH=${TARGETARCH}) /usr/bin/nsh
  ENTRYPOINT ["/usr/bin/nsh"]
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
  RUN CGO_ENABLED=0 go build -o nsh --ldflags "-s \
    -X 'github.com/noisysockets/nsh/internal/constants.Version=${VERSION}'"
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