COVER=go tool cover

all: install

install:
	go install github.com/harrybrwn/apizza

uninstall:
	$(RM) "$$GOPATH/bin/apizza"

build:
	go build -o bin/apizza

release:
	bash scripts/release.sh

test: coverage.txt test-build
	bash scripts/integration.sh ./bin/apizza
	@[ -d bin ] && rm -rf bin

test-build:
	go build -o bin/apizza -ldflags "-X cmd.enableLog=false"

coverage.txt:
	bash scripts/test.sh

html: coverage.txt
	$(COVER) -html=$<

clean:
	$(RM) coverage.txt
	$(RM) -r release bin
	go clean -testcache

.PHONY: install test clean html release