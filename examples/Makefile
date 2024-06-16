OUTPUT_DIR = ./output
GIT_COMMIT = $(shell git rev-parse HEAD | cut -c1-7)
GIT_TAG = $(shell git describe --tags HEAD)
VERSION = $(shell echo $(GIT_TAG) | tr -d 'v')
PACKAGE = github.com/lazygo/lazygo/examples
BUILD_OPTIONS = -ldflags "-w -s -X $(PACKAGE)/framework.Version=$(VERSION) -X $(PACKAGE)/framework.BuildID=$(GIT_COMMIT)" -trimpath

# GO_PROXY = GOPROXY=https://mirrors.aliyun.com/goproxy/

default: all

all: deps fmt build

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: fmt
fmt:
	go fmt ./

.PHONY: build
build:
	CGO_ENABLED=0 ${GO_PROXY} go build -o ${OUTPUT_DIR}/app ${BUILD_OPTIONS}

.PHONY: docker
docker: deps vendor
	docker build -t ${PACKAGE}:${GIT_TAG} .
	docker push ${PACKAGE}:${GIT_TAG}

.PHONY: run
run: deps
	go run ${BUILD_OPTIONS} main.go

.PHONY: release
release:
	git pull && \
	last_tag=`git tag --list | grep -E '(\d+\.){1,3}' | tail -n 1 | \
	awk -F. -v OFS=. 'NF==1{print ++$$NF}; NF>1{if(length($$NF+1)>length($$NF))$$(NF-1)++; $$NF=sprintf("%0*d", length($$NF), ($$NF+1)%(10^length($$NF))); print}'` && \
	git tag $$last_tag && git push --tag
