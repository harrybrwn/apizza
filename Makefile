PKGS=./cmd ./dawg ./pkg/cache ./pkg/config
COVER_FILE=test-coverage
COVER=go tool cover


build: test
	go install apizza

test:
	go test -v ./... -coverprofile=$(COVER_FILE)
	$(COVER) -func=$(COVER_FILE)
	$(COVER) -html=$(COVER_FILE) -o coverage.html

clean: build
	rm $(COVER_FILE)