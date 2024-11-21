package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"weaviate/engine.go"
	"weaviate/storage"
)

const DefaultSegmentDir = "segment-data"

func main() {
	dir := flag.String("dir", DefaultSegmentDir, "Directory to load segment files from")
	flag.Parse()

	files, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", *dir, err)
		return
	}

	var segments []*storage.Segment
	var totalDocs uint32
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".bin" {
			continue
		}
		segmentPath := filepath.Join(*dir, file.Name())
		segmentFile, err := os.Open(segmentPath)
		if err != nil {
			fmt.Printf("Error opening segment file %s: %v\n", segmentPath, err)
			continue
		}
		defer segmentFile.Close()

		segment := storage.NewSegment()
		segment.Deserialize(segmentFile)
		totalDocs += segment.TotalDocs()
		segments = append(segments, segment)
		segment.PrintInfo()
	}

	fmt.Printf("Loaded %d segments\n", len(segments))

	queryEngine, err := engine.NewQueryEngine(segments, totalDocs)
	if err != nil {
		panic(err)
	}

	query, exists := os.LookupEnv("QUERY")
	if !exists {
		query = "great vector database"
	}

	terms := strings.Fields(query)
	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Terms: %v\n", terms)
	scoredDocuments, err := queryEngine.MultiTermQuery(terms, func(doc1, doc2 engine.ScoredDocument) bool {
		return doc1.Score > doc2.Score
	})
	if err != nil {
		fmt.Printf("No result: %v\n", err)
		return
	}

	fmt.Printf("Scored documents: %d\n", len(scoredDocuments))
	fmt.Println(strings.Repeat("-", 22))
	fmt.Printf("| %-8s | %-8s |\n", "DocID", "Score")
	fmt.Println(strings.Repeat("-", 22))
	for _, scoredDocument := range scoredDocuments {
		fmt.Printf("| %-8d | %8.2f |\n", scoredDocument.DocID, scoredDocument.Score)
	}
	fmt.Println(strings.Repeat("-", 22))

}
