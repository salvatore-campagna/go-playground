.DEFAULT_GOAL := build

BINARY_NAME = collatz-conjecture

.PHONY: fmt vet build run clean

build: vet
	go build -o $(BINARY_NAME) ./

run: build
	./$(BINARY_NAME)

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean:
	rm -f $(BINARY_NAME)
