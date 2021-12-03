.PHONY: all
.PHONY: build
.PHONY: test

all: build

resources/data.go: content/*
	go run tools/resourceStorageBuilder.go -p content:content/ -d resources

build: build/qbert-linux-x86

go.sum:
	go mod tidy

build/qbert-linux-x86: go.sum resources/data.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o build/qbert-linux-x86 *.go

clean:
	rm -f go.sum resources/data.go build/qbert-linux-x86
