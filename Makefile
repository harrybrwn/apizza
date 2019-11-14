COVER_FILE=test-coverage
COVER=go tool cover

all: install

install:
	go install

release:
	bash scripts/release.sh

test: setup $(COVER_FILE)
	go test -cover ./... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)

html: test
	$(COVER) -html=$(COVER_FILE)

setup:
	touch $(COVER_FILE)

clean:
	$(RM) $(COVER_FILE)
	$(RM) -r release

.PHONY: install test setup clean html release