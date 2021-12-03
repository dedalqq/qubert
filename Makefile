.PHONY: all
.PHONY: build
.PHONY: test

COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --tags)
LDFLAGS := "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -s -w"

all: build

resources/data.go: content/*
	go run tools/resourceStorageBuilder.go -p content:content/ -d resources

build: build/qubert-linux-x86-64 build/qubert-linux-arm64

go.sum:
	go mod tidy

build/qubert-linux-x86-64: go.sum resources/data.go
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o build/qubert-linux-x86-64 *.go

build/qubert-linux-arm64: go.sum resources/data.go
	GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o build/qubert-linux-arm64 *.go

clean:
	rm -f go.sum resources/data.go build/qubert-linux-x86
