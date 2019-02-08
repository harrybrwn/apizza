COVER=test-coverage

test:
	go test ./... -coverprofile=$(COVER)
	go tool cover -func=$(COVER)
	go tool cover -html=$(COVER) -o coverage.html

