COVER=go tool cover

VERSION=$(shell git describe --tags --abbrev=12)
GOFLAGS=-ldflags "-X $(shell go list)/cmd.version=$(VERSION)"

build:
	go build $(GOFLAGS)

install:
	go install $(GOFLAGS)

uninstall:
	go clean -i

test: test-build
	bash scripts/test.sh
	bash scripts/integration.sh ./bin/apizza
	@[ -d ./bin ] && [ -x ./bin/apizza ] && rm -rf ./bin

release:
	scripts/release build

test-build:
	scripts/build.sh test

coverage.txt:
	@ echo '' > coverage.txt
	go test -v ./... -coverprofile=coverage.txt -covermode=atomic

html: coverage.txt
	$(COVER) -html=$<

clean:
	$(RM) -r coverage.txt release/apizza-* bin
	go clean -testcache
	go clean

all: test build release

.PHONY: install test clean html release
