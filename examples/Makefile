OUTPUT_DIR = ./output
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
GIT_TAG = `git describe --tags HEAD`
VERSION = $(GIT_TAG)
BUILD_OPTIONS = -ldflags "-linkmode external -extldflags -w -s -X main.Version=$(VERSION) -X main.BuildID=$(GIT_COMMIT)" -trimpath

GO_PROXY = GOPROXY=https://mirrors.aliyun.com/goproxy/

default: all

all: deps fmt build

.PHONY: deps
deps:
	GOPROXY=https://mirrors.aliyun.com/goproxy/ go mod tidy && go mod download

.PHONY: fmt
fmt:
	go fmt ./

.PHONY: build
build:
	CGO_ENABLED=0 ${GO_PROXY} go build -o ${OUTPUT_DIR}/app ${BUILD_OPTIONS}

.PHONY: docker
docker:
	docker build -t ccr.ccs.tencentyun.com/net51/gohudong:${GIT_TAG} .
	docker push ccr.ccs.tencentyun.com/net51/gohudong:${GIT_TAG}

.PHONY: run
run: deps
	go run ${BUILD_OPTIONS} main.go

.PHONY: release
release:
	git pull && \
	last_tag=`git tag --list | grep -E '(\d+\.){1,3}' | tail -n 1 | \
	awk -F. -v OFS=. 'NF==1{print ++$$NF}; NF>1{if(length($$NF+1)>length($$NF))$$(NF-1)++; $$NF=sprintf("%0*d", length($$NF), ($$NF+1)%(10^length($$NF))); print}'` && \
	git tag $$last_tag && git push --tag

