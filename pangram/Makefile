.DEFAULT_GOAL := test
.PHONY: fmt vet test clean docs

test: vet
	go test

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean:

