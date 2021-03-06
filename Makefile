BINARY := emoji-generator
BUILDFLAGS := "-s -w"

.PHONY: build release

default: build

build:
	mkdir -p builds
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(BUILDFLAGS) -o $(BINARY)-linux *.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(BUILDFLAGS) -o $(BINARY)-darwin *.go
