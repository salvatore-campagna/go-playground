.DEFAULT_GOAL := run
.PHONY: fmt vet test clean

test: build
	go test ./...

run: build
	go run main/main.go

build: vet
	go build ./...

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean:
	go clean
