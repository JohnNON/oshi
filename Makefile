all: lint test

.PHONY:
test:
	go test -v ./...

.PHONY:
lint:
	golangci-lint run
