# General

CURRENT_OS = $(shell uname | tr A-Z a-z)
EXECUTABLE_EXTENSION = 
WORKDIR = $(PWD)

# Go parameters
GOCMD = go
GOTEST = $(GOCMD) test -v

# Coverage
COVERAGE_REPORT = coverage.txt
COVERAGE_MODE = atomic
NAME=torrentd
AUTHOR=sp0x
ARCH=amd64
OS=linux darwin windows

#Dependency versions

GOLANGCI_VERSION = 1.33.0

ifeq ($(suffix $(SHELL)),.exe)
    # Windows system
	EXECUTABLE_EXTENSION := .exe
endif

ifneq ($(origin CI), undefined)
	WORKDIR := $(GOPATH)/src/github.com/$(AUTHOR)/$(NAME)
endif

build:
	go build -o $(NAME)$(EXECUTABLE_EXTENSION) ./cmd

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ./bin/ v${GOLANGCI_VERSION}
	@mv bin/golangci-lint "$@"

golint:
	golint ./...

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	bin/golangci-lint run

.PHONY: fix
fix: bin/golangci-lint ## Fix lint violations
	bin/golangci-lint run --fix

#Note that gox is required for multi-arch build
build-multi-arch:
	gox -os="${OS}" -arch="${ARCH}" -output="$(NAME).{{.OS}}.{{.Arch}}" -ldflags "-s -w -X main.Rev=`git rev-parse --short HEAD`" -verbose ./...

swagger: install-deps
	swag init --dir ./,./cmd,./server -g ./cmd/serve.go

assets: install-deps swagger
	@echo "Embedding assets as code"
	bindata -o indexer/definitions/assets.go ./definitions/...


# go get github.com/konsorten/go-windows-terminal-sequences
# go get github.com/inconshreveable/mousetrap
install-deps:
	@echo "Installing go utils"
	go install github.com/swaggo/swag/cmd/swag@latest
	go get github.com/kataras/bindata/cmd/bindata
	go install github.com/mitchellh/gox@latest

install:
	go build -i -o $(GOPATH)/bin/$(NAME) ./cmd

test:
	@cd $(WORKDIR); \
	$(GOTEST) ./...

test-coverage:
	@cd $(WORKDIR); \
	echo "" > $(COVERAGE_REPORT); \
	$(GOTEST) -failfast -coverprofile=$(COVERAGE_REPORT).tmp -coverpkg=./... -covermode=$(COVERAGE_MODE) ./...; \
	cat $(COVERAGE_REPORT).tmp | grep -v "indexer/definitions/assets.go" | grep -v "/mocks/" > $(COVERAGE_REPORT); \
	rm $(COVERAGE_REPORT).tmp

build-image:
	docker-compose build torrentd
	docker-compose push

benchmark-server:
	#go-wrk -redir -d 20  -c 100 http://127.0.0.1:5000/status
	go-wrk -redir -T 2000 -d 1 -c 1 "http://127.0.0.1:5000/torznab/all/api?t=search&cat=2000,2010,2020,2030,2035,2040,2045,2050,2060&extended=1&apikey=210fc7bb818639a&offset=0&limit=100&q=accountant 2016"