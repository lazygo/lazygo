FROM golang:1.17-alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories & apk add --no-cache --update git gcc rust make
COPY ../../../Users/Administrator/Desktop/gohudong /src
WORKDIR /src

RUN go version && make build

FROM alpine:3.15

COPY --from=builder /src/output/app /usr/local/bin/app

ENTRYPOINT ["/usr/local/bin/app"]