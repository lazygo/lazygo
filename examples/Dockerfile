FROM golang:1.26-alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories & apk add --no-cache --update git gcc rust make
COPY . /src
WORKDIR /src

ARG CGO_ENABLED=0
RUN go version && CGO_ENABLED=$CGO_ENABLED make build

FROM alpine:3.20

COPY --from=builder /src/output/app /usr/local/bin/app

ENTRYPOINT ["/usr/local/bin/app"]