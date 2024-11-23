.DEFAULT_GOAL := help

.PHONY: all build clean dataclean datagen stats index query test fmt vet help

BIN_DIR := bin
CREATE_INDEX := $(BIN_DIR)/index
QUERY_INDEX := $(BIN_DIR)/query
DATA_GEN := $(BIN_DIR)/datagen
STATS := $(BIN_DIR)/stats
DATA_CLEAN := $(BIN_DIR)/dataclean

JSON_FILE := segments.json
SEGMENT_DIR := segment-data
STATS_FILE := stats.txt
DATA_CLEAN_INPUT_FILE := data/segments.json
DATA_CLEAN_OUTPUT_FILE := segments.json

build: $(CREATE_INDEX) $(QUERY_INDEX) $(DATA_GEN) $(STATS) $(DATA_CLEAN) test

$(CREATE_INDEX): clean cmd/index/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(CREATE_INDEX) cmd/index/main.go

$(QUERY_INDEX): clean cmd/query/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(QUERY_INDEX) cmd/query/main.go

$(DATA_GEN): clean cmd/datagen/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(DATA_GEN) cmd/datagen/main.go

$(STATS): clean cmd/stats/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(STATS) cmd/stats/main.go

$(DATA_CLEAN): clean cmd/dataclean/main.go
	@mkdir -p $(BIN_DIR)
	go build -o $(DATA_CLEAN) cmd/dataclean/main.go

test: vet
	go test ./...

test-verbose:
	go test -v ./...

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

delete:
	rm -rf $(SEGMENT_DIR)

index: $(CREATE_INDEX)
	./$(CREATE_INDEX) -path $(JSON_FILE) -dir $(SEGMENT_DIR)

query: $(QUERY_INDEX)
	./$(QUERY_INDEX) -dir $(SEGMENT_DIR)

datagen: $(DATA_GEN)
	./$(DATA_GEN) -path $(JSON_FILE)

stats: $(STATS)
	./$(STATS) -path $(JSON_FILE) > $(STATS_FILE)

dataclean: $(DATA_CLEAN)
	./$(DATA_CLEAN) -input $(DATA_CLEAN_INPUT_FILE) -output $(DATA_CLEAN_OUTPUT_FILE)

clean:
	rm -rf $(BIN_DIR)

help:
	@echo "Makefile for building and running executables"
	@echo "Usage:"
	@echo "  make build                - Build all executables"
	@echo "  make index                - Run the create-index executable"
	@echo "  make query                - Run the query-index executable"
	@echo "  make datagen              - Run the data-gen executable"
	@echo "  make stats                - Run the stats executable"
	@echo "  make clean                - Remove all built files"
	@echo "  make test                 - Run tests"
	@echo "  make fmt                  - Format Go code"
	@echo "  make vet                  - Run go vet"
	@echo ""
	@echo "Optional parameters (set using 'make param=value'):"
	@echo "  JSON_FILE                 - Path to the data generation input file (default: segments_clean.json)"
	@echo "  SEGMENT_DIR               - Directory for segment data (default: segment-data)"
	@echo "  STATS_FILE                - Path to the statistics output file (default: stats_output.txt)"
	@echo "  DATA_CLEAN_INPUT_FILE     - Path to the data clean input file (default: segments.json)"
	@echo "  DATA_CLEAN_OUTPUT_FILE    - Path to the data clean output file (default: segments_clean.json)"
