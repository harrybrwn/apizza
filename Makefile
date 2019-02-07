test:
	go test ./... -coverprofile=test-coverage
	go tool cover -func=test-coverage

build:
	echo "this work'in?"
