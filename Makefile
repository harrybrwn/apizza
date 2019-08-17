COVER_FILE=test-coverage
COVER=go tool cover
PKG=github.com/harrybrwn/apizza

all: test install clean

install:
	go install $(PKG)

test: setup $(COVER_FILE)
	go test -cover $(PKG)/... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)

html: test
	$(COVER) -html=$(COVER_FILE)

setup:
	touch $(COVER_FILE)

clean:
	rm $(COVER_FILE)

.PHONY: install test setup clean html