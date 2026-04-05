OUTPUT_DIR = ./output
GIT_COMMIT = $(shell git rev-parse HEAD | cut -c1-7)
GIT_TAG = $(shell git describe --tags HEAD)
VERSION = $(shell echo $(GIT_TAG) | tr -d 'v')
PACKAGE = github.com/lazygo/lazygo/examples
CONTAINER ?= lazygo-examples
DOCKER_IMAGE = ${PACKAGE}/${CONTAINER}

ifeq ($(shell uname), Darwin)
	EXTRA_LDFLAGS =
else
	EXTRA_LDFLAGS = -extldflags='-static'
endif
BUILD_OPTIONS = -buildvcs=false -ldflags "-w -s -X $(PACKAGE)/framework.Version=$(VERSION) -X $(PACKAGE)/framework.BuildID=$(GIT_COMMIT) $(EXTRA_LDFLAGS)" -trimpath

# 默认关闭 CGO；需要时用 CGO_ENABLED=1 make build 或导出环境变量覆盖
CGO_ENABLED ?= 0
WINDOWS_CC ?= $(WINDOWS_CC)
LINUX_CC ?= $(LINUX_CC)
DARWIN_CC ?= $(DARWIN_CC)

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
	@if [ "$(CGO_ENABLED)" = "1" ]; then \
		go install github.com/goware/modvendor@latest ; \
		modvendor -copy="**/*.c **/*.h **/*.a" -v ; \
	fi

.PHONY: fmt
fmt:
	go fmt ./

.PHONY: build
build:
	# 默认 CGO_ENABLED=0；开启示例: CGO_ENABLED=1 make build
	CGO_ENABLED=$(CGO_ENABLED) ${GO_PROXY} go build -o ${OUTPUT_DIR}/app ${BUILD_OPTIONS}

.PHONY: docker
docker: deps vendor
	docker build -t ${DOCKER_IMAGE}:${GIT_TAG} .
	rm -rf vendor
	docker push ${DOCKER_IMAGE}:${GIT_TAG}

.PHONY: run
run: deps
	go run ${BUILD_OPTIONS} main.go

.PHONY: release
release:
	git pull
	git push
	last_tag=`git tag --list | grep -E 'v?([0-9]+\.){1,3}[0-9]' | sort -V | tail -n 1 | tr -d 'v' | \
	awk -F. -v OFS=. 'NF==1{print ++$$NF}; NF>1{ $$NF=$$NF+1; for (i=NF; i>1; i--) { if ($$i>=100) { $$i=0; $$(i-1)++ }} print}'` && \
	echo tag: v$$last_tag && \
	git tag v$$last_tag && git push --tag
