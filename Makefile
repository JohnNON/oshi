.PHONY:
test:
	go test -v ./...

.PHONY:
lint:
	golangci-lint run
