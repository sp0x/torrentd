# General
WORKDIR = $(PWD)

# Go parameters
GOCMD = go
GOTEST = $(GOCMD) test -v

# Coverage
COVERAGE_REPORT = coverage.txt
COVERAGE_MODE = atomic
NAME=torrent-rss
ARCH=amd64
OS=linux darwin windows

ifneq ($(origin CI), undefined)
	WORKDIR := $(GOPATH)/src/github.com/$(NAME)
endif

build:
	gox -os="${OS}" -arch="${ARCH}" -output="$(NAME).{{.OS}}.{{.Arch}}" -ldflags "-s -w -X main.Rev=`git rev-parse --short HEAD`" -verbose ./...

install:
	go build -i -o $(GOPATH)/bin/$(NAME) ./cmd

test:
	@cd $(WORKDIR); \
	$(GOTEST) ./...

test-coverage:
	@cd $(WORKDIR); \
	echo "" > $(COVERAGE_REPORT); \
	$(GOTEST) -coverprofile=$(COVERAGE_REPORT) -coverpkg=./... -covermode=$(COVERAGE_MODE) ./...
