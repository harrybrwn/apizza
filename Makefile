COVER_FILE=test-coverage
COVER=go tool cover

all: test install clean

.PHONY: install test setup clean

install:
	go install github.com/harrybrwn/apizza

test: setup
	go test -cover ./... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)

setup:
	touch $(COVER_FILE)

clean:
	rm $(COVER_FILE)