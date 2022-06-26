OUTPUT_DIR = ./builds
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
VERSION = 0.0.2
BUILD_OPTIONS = -ldflags "-w -s -X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)" -trimpath

default: all

all: deps fmt build

.PHONY: deps
deps:
	go mod tidy && go mod download

.PHONY: fmt
fmt:
	go fmt ./

.PHONY: build
build:
	go build ${BUILD_OPTIONS}

.PHONY: dev
dev:
	go run ${BUILD_OPTIONS} main.go
