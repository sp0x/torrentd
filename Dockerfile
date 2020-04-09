FROM golang:1.13-alpine AS builder
LABEL maintainer="vaskovasilev94@yahoo.com"

RUN apk update && apk add --no-cache git  gcc musl-dev
ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .
RUN go build -o /go/bin/torrent-rss ./cmd


FROM scratch
COPY --from=builder /go/bin/torrent-rss /go/bin/torrent-rss

ENTRYPOINT ["/go/bin/torrent-rss"]