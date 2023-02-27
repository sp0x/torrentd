FROM golang:1.20-alpine AS builder
LABEL maintainer="vaskovasilev94@yahoo.com"

RUN apk update && apk add --no-cache git  gcc musl-dev curl
ENV GO111MODULE=on

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .
RUN go build -o /go/bin/torrentd -ldflags="-w -s" ./cmd
# Optional: in case your application uses dynamic linking (often the case with CGO),
# this will collect dependent libraries so they're later copied to the final image
# NOTE: make sure you honor the license terms of the libraries you copy and distribute
WORKDIR /dist
RUN cp /go/bin/torrentd .
RUN ldd ./torrentd | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
    
FROM alpine:latest as certs
RUN apk --update add ca-certificates

#Finish the build
FROM scratch
ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /dist /
EXPOSE 5000
ENV GIN_MODE=release
ENTRYPOINT ["/torrentd"]