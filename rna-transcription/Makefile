.DEFAULT_GOAL := test
.PHONY: fmt vet test clean

test: vet
	go test

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean:
	go clean
 