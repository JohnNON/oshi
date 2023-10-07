.PHONY:
test:
	go test ./oshi_test

.PHONY:
lint:
	golangci-lint run
