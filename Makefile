COVER=go tool cover

all: install

install:
	go install github.com/harrybrwn/apizza

build:
	go build -o bin/apizza

release:
	bash scripts/release.sh

test: build coverage.txt
	bash scripts/integration.sh ./bin/apizza
	@[ -d bin ] && rm -rf bin

coverage.txt:
	bash scripts/test.sh

html: coverage.txt
	$(COVER) -html=$<

clean:
	$(RM) coverage.txt
	$(RM) -r release bin

.PHONY: install test setup clean html release