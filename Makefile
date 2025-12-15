BINARY_NAME=ghex
VERSION=$(shell cat VERSION 2>/dev/null || echo "1.0.0")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build build-all clean test run

all: build

build:
	go build $(LDFLAGS) -o build/$(BINARY_NAME) ./cmd/ghex

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-amd64 ./cmd/ghex

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-arm64 ./cmd/ghex

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-amd64 ./cmd/ghex

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-arm64 ./cmd/ghex

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-windows-amd64.exe ./cmd/ghex

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

clean:
	rm -rf build/

test:
	go test -v ./...

test-prop:
	go test -v -run "Prop" ./...

run:
	go run ./cmd/ghex

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

checksums:
	cd build && sha256sum $(BINARY_NAME)-* > checksums.txt

install: build
	cp build/$(BINARY_NAME) /usr/local/bin/

uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)
