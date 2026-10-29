NAME = serenia
TAG = $(shell git describe --tags --always)
VERSION = $(TAG:v%=%)

PARAMS = -v -trimpath -ldflags "-s -w -buildid="
MAIN_PARAMS = $(PARAMS)
MAIN = ./cmd/serenia

ARTIFACT_DIR = ./build
ARTIFACT = $(ARTIFACT_DIR)/$(NAME)

build:
	mkdir -p build
	go build $(MAIN_PARAMS) -o $(ARTIFACT) $(MAIN)

lint:
	GOOS=linux golangci-lint run ./...
	GOOS=freebsd golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -rf $(ARTIFACT_DIR)

