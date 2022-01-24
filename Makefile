COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --tags)
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -s -w

.PHONY: all build

DEPENDS := go.sum \
	resources/data.go \
	config.go \
	main.go \
	application/*.go \
	logger/*.go \
	plugins/*/*.go \
	resources/storage.go \
	tools/*.go \
	uuid/*.go \
	pluginTools/*.go

all: build

resources/data.go: content/*
	go run tools/resourceStorageBuilder.go -p content:content/ -d resources

build: build/qubert-linux-x86-64 build/qubert-linux-arm64

go.sum:
	go mod tidy

build/qubert-linux-x86-64: $(DEPENDS)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o build/qubert-linux-x86-64 *.go

build/qubert-linux-arm64: $(DEPENDS)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o build/qubert-linux-arm64 *.go

clean:
	rm -f go.sum resources/data.go build/qubert*
