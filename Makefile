COVER=go tool cover

test: test-build
	bash scripts/test.sh
	bash scripts/integration.sh ./bin/apizza
	@[ -d ./bin ] && [ -x ./bin/apizza ] && rm -rf ./bin

install:
	go install github.com/harrybrwn/apizza

uninstall:
	$(RM) "$$GOPATH/bin/apizza"

build:
	go build -o bin/apizza

release:
	scripts/release build

test-build:
	go build -o bin/apizza -ldflags "-X cmd.enableLog=false"

coverage.txt:
	@ echo '' > coverage.txt
	go test -v ./... -coverprofile=coverage.txt -covermode=atomic

html: coverage.txt
	$(COVER) -html=$<

clean:
	$(RM) coverage.txt release/apizza-linux release/apizza-windows release/apizza-darwin
	$(RM) -r bin
	go clean -testcache

all: test build release

.PHONY: install test clean html release
