package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"weaviate/fetcher"
	"weaviate/storage"
)

const (
	DefaultSegmentDir = "segment-data"
	MaxDocsPerSegment = 1_000_000
)

func main() {
	jsonFilePath := flag.String("path", "", "Path to the input JSON file")
	dir := flag.String("dir", DefaultSegmentDir, "Directory to store segment files")
	flag.Parse()
	fmt.Printf("Reading file: %s\n", *jsonFilePath)

	if err := os.MkdirAll(*dir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", *dir, err)
		return
	}

	data, err := fetcher.FetchJson(*jsonFilePath)
	if err != nil {
		fmt.Printf("Error fetching JSON: %v\n", err)
		return
	}

	jsonSegments, err := fetcher.ParseTermPostings(data)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	fmt.Printf("Processing %d segments\n", len(jsonSegments))

	segments := make([]*storage.Segment, 0)
	for segmentID, jsonDocuments := range jsonSegments {
		segment := storage.NewSegment()
		segments = append(segments, segment)
		segment.BulkIndex(jsonDocuments)

		segmentPath := filepath.Join(*dir, fmt.Sprintf("segment_%d.bin", segmentID))
		segmentFile, err := os.Create(segmentPath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", segmentPath, err)
			return
		}

		if err := segment.WriteSegment(segmentFile); err != nil {
			fmt.Printf("Error writing segment %s: %v\n", segmentPath, err)
			segmentFile.Close()
			return
		}
		segmentFile.Close()
	}

	fmt.Println("Segments created successfully.")
}
