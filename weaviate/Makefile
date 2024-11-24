.DEFAULT_GOAL := help

.PHONY: all build clean index query test fmt vet help download archive doc

BIN_DIR := bin
CREATE_INDEX := $(BIN_DIR)/index
QUERY_INDEX := $(BIN_DIR)/query

JSON_FILE := segments.json
SEGMENT_DIR := segment-data
DATA_URL := https://storage.googleapis.com/weaviate-tech-challenges/db-engineer/segments.json
ARCHIVE_NAME := project.zip

build: $(CREATE_INDEX) $(QUERY_INDEX) test

$(CREATE_INDEX): clean cmd/index/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(CREATE_INDEX) cmd/index/main.go

$(QUERY_INDEX): clean cmd/query/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(QUERY_INDEX) cmd/query/main.go

test: vet
	go test ./...

test-verbose:
	go test -v ./...

test-storage:
	go test ./storage

test-fetcher:
	go test ./fetcher

test-engine:
	go test ./engine

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

delete:
	rm -rf $(SEGMENT_DIR)

index: $(CREATE_INDEX) clean-segments
	./$(CREATE_INDEX) -path $(JSON_FILE) -dir $(SEGMENT_DIR)

query: $(QUERY_INDEX)
	./$(QUERY_INDEX) -dir $(SEGMENT_DIR) -query "$(QUERY)"

download:
	@echo "Downloading data from $(DATA_URL) to $(JSON_FILE)..."
	curl -XGET -o $(JSON_FILE) "$(DATA_URL)"
	@echo "Download completed. Data saved to $(JSON_FILE)."

archive:
	@echo "Creating archive: $(ARCHIVE_NAME)"
	bash scripts/archive.sh
	@echo "Archive created successfully: $(ARCHIVE_NAME)"

doc:
	@echo "Starting Go documentation server..."
	godoc -http=:6060 &
	sleep 10
	@echo "Opening Go documentation in browser..."
	xdg-open http://localhost:6060 >/dev/null 2>&1 || open http://localhost:6060 || echo "Please open http://localhost:6060 in your browser."

clean:
	rm -rf $(BIN_DIR)

clean-segments:
	rm -rf $(SEGMENT_DIR)

help:
	@echo "Makefile for building and running executables"
	@echo "Usage:"
	@echo "  make build                 - Build all executables"
	@echo "  make index                 - Run the index executable"
	@echo "                               Example: ./$(CREATE_INDEX) -path $(JSON_FILE) -dir $(SEGMENT_DIR)"
	@echo "  make query [QUERY=<terms>] - Run the query executable with optional terms"
	@echo "                               Example (custom query): make query QUERY='great database'"
	@echo "  make download              - Download data from the provided URL"
	@echo "  make archive               - Create a zip archive of the project"
	@echo "  make doc                   - Start the Go documentation server and open it in a browser"
	@echo "  make clean                 - Remove all built files"
	@echo "  make test                  - Run tests"
	@echo "  make fmt                   - Format Go code"
	@echo "  make vet                   - Run go vet"
