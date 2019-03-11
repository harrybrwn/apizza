COVER_FILE=test-coverage
COVER=go tool cover
OUTPUT=$(GOPATH)/bin/apizza$(GOEXE)

all: test build clean

.PHONY: build build-win test setup clean reset

build:
	go install apizza

build-win:
	GOOS=windows go build -o=$(OUTPUT).exe apizza

build-mac:
	GOOS=darwin go build -o $(OUTPUT) apizza

test: setup
	go test -cover ./... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)
	$(COVER) -html=$(COVER_FILE) -o coverage.html

setup:
	touch $(COVER_FILE)

clean:
	rm $(COVER_FILE)

reset:
	rm $(HOME).apizza/cache/apizza.db
