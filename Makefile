COVER_FILE=test-coverage
COVER=go tool cover

all: test build clean

.PHONY: build test setup clean

build:
	go install github.com/harrybrwn/apizza

test: setup
	go test -cover ./... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)
	$(COVER) -html=$(COVER_FILE) -o coverage.html

setup:
	touch $(COVER_FILE)

clean:
	rm $(COVER_FILE)