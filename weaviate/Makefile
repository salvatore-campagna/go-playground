.DEFAULT_GOAL := help

.PHONY: all build clean run-create-index run-query-index run-data-gen test fmt vet

BIN_DIR := bin
CREATE_INDEX := $(BIN_DIR)/create-index
QUERY_INDEX := $(BIN_DIR)/query-index
DATA_GEN := $(BIN_DIR)/data-gen

# Build all executables
build: $(CREATE_INDEX) $(QUERY_INDEX) $(DATA_GEN) test

# Build create-index executable
$(CREATE_INDEX): clean cmd/create-index/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(CREATE_INDEX) cmd/create-index/main.go

# Build query-index executable
$(QUERY_INDEX): clean cmd/query-index/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(QUERY_INDEX) cmd/query-index/main.go

# Build data-gen executable
$(DATA_GEN): clean cmd/data-gen/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(DATA_GEN) cmd/data-gen/main.go

test: vet
	go test ./...

test-verbose:
	go test -v ./...

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

delete:
	rm -rf segment-data
	rm -rf

# Run create-index executable
run-create-index: $(CREATE_INDEX)
	./$(CREATE_INDEX) -dir segment-data

# Run query-index executable
run-query-index: $(QUERY_INDEX)
	./$(QUERY_INDEX) -dir segment-data

# Run data-gen executable
run-data-gen: $(DATA_GEN)
	./$(DATA_GEN) -output generated_segments.json

# Clean up built executables
clean:
	rm -rf $(BIN_DIR)

# Display help
help:
	@echo "Makefile for building and running executables"
	@echo "Usage:"
	@echo "  make build                - Build all executables"
	@echo "  make run-create-index     - Run the create-index executable"
	@echo "  make run-query-index      - Run the query-index executable"
	@echo "  make run-data-gen         - Run the data-gen executable"
	@echo "  make clean                - Remove all built files"
	@echo "  make test                 - Run tests"
	@echo "  make fmt                  - Format Go code"
	@echo "  make vet                  - Run go vet"

