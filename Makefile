BINARY_NAME=multivpn
VERSION=$(shell cat VERSION 2>/dev/null || echo "1.0.0")
GIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "release")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitHash=$(GIT_HASH) -s -w"

.PHONY: all build run test clean build-windows

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

run: build
	./$(BINARY_NAME) web

test:
	go test -v ./...

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME).exe main.go

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
