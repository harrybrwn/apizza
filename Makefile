COVER=go tool cover

all: install

install:
	go install github.com/harrybrwn/apizza

build:
	go build -o bin/apizza

release:
	bash scripts/release.sh

test: coverage.txt test-build
	bash scripts/integration.sh ./bin/apizza
	@[ -d bin ] && rm -rf bin

test-build:
	go build -o bin/apizza -ldflags "-X cmd.enableLog=false -X cmd.Logger=ioutil.Discard"

coverage.txt:
	bash scripts/test.sh

html: coverage.txt
	$(COVER) -html=$<

clean:
	$(RM) coverage.txt
	$(RM) -r release bin

.PHONY: install test clean html release