COVER_FILE=test-cover
COVER=go tool cover

all: install

install:
	go install github.com/harrybrwn/apizza

release:
	bash scripts/release.sh

test: setup $(COVER_FILE)
	go test -cover ./... -coverprofile=$(COVER_FILE) -covermode=atomic
	$(COVER) -func=$(COVER_FILE)

html: test
	$(COVER) -html=$(COVER_FILE)

setup:
	touch $(COVER_FILE)

clean:
	$(RM) $(COVER_FILE) coverage.txt
	$(RM) -r release

.PHONY: install test setup clean html release