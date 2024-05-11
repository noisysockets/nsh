VERSION 0.8
FROM golang:1.22-bookworm
WORKDIR /workspace

all:
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

build:
  ARG GOOS=linux
  ARG GOARCH=amd64
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  COPY +build-web/dist ./web/dist
  RUN CGO_ENABLED=0 go build --ldflags '-s' -o nsh main.go
  SAVE ARTIFACT ./nsh AS LOCAL dist/nsh-${GOOS}-${GOARCH}

tidy:
  BUILD +tidy-go
  BUILD +tidy-web

lint:
  BUILD +lint-go
  BUILD +lint-web

test:
  BUILD +test-go
  BUILD +test-web

tidy-go:
  LOCALLY
  RUN go mod tidy
  RUN go fmt ./...

lint-go:
  FROM golangci/golangci-lint:v1.57.2
  WORKDIR /workspace
  COPY . ./
  RUN mkdir -p web/dist \
    && echo '<html></html>' > web/dist/index.html
  RUN golangci-lint run --timeout 5m ./...

test-go:
  COPY . ./
  RUN mkdir -p web/dist \
    && echo '<html></html>' > web/dist/index.html
  RUN go test -coverprofile=coverage.out -v ./...
  SAVE ARTIFACT ./coverage.out AS LOCAL coverage.out

tidy-web:
  FROM +deps-web
  COPY web .
  RUN npm run format
  SAVE ARTIFACT src AS LOCAL web/src

lint-web:
  FROM +deps-web
  COPY web .
  RUN npm run lint

test-web:
  FROM +deps-web
  COPY web .
  RUN npm run test

build-web:
  FROM +deps-web
  COPY web .
  RUN npm run build
  SAVE ARTIFACT dist AS LOCAL web/dist

deps-web:
  FROM +tools
  COPY web/package.json web/package-lock.json ./
  RUN npm install

tools:
  RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
  RUN apt install -y nodejs 