COVER_FILE=test-cover
COVER=go tool cover

all: install

install:
	go install github.com/harrybrwn/apizza

release:
	bash scripts/release.sh

test: setup $(COVER_FILE)
	go test -cover ./... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)

install-deps:
	go list -f '{{ join .Imports "\n" }}' ./... | \
		grep -P '^(github.com|gopkg.in)/.*' | \
		grep -v "`go list`" | \
		awk '{print}' ORS=' ' | \
		go get -u


html: test
	$(COVER) -html=$(COVER_FILE)

setup:
	touch $(COVER_FILE)

clean:
	$(RM) $(COVER_FILE)
	$(RM) -r release

.PHONY: install test setup clean html release