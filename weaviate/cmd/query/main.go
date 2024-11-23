package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"weaviate/engine"
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
		segment := storage.NewSegment()

		if err := loadSegment(segmentPath, segment); err != nil {
			fmt.Printf("Error loading segment %s: %v\n", segmentPath, err)
			continue
		}

		totalDocs += segment.TotalDocs()
		segments = append(segments, segment)
		segment.PrintInfo()
	}

	if len(segments) == 0 {
		fmt.Println("No valid segments found.")
		return
	}

	queryEngine, err := engine.NewQueryEngine(segments, totalDocs)
	if err != nil {
		panic(err)
	}

	query := getQuery()
	terms := strings.Fields(query)

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Terms: %v\n", terms)

	scoredDocuments, err := queryEngine.MultiTermQuery(terms, func(doc1, doc2 engine.ScoredDocument) bool {
		return doc1.Score > doc2.Score
	})
	if err != nil {
		fmt.Printf("Query execution failed: %v\n", err)
		return
	}

	printResults(scoredDocuments)
}

func loadSegment(path string, segment *storage.Segment) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return segment.Deserialize(file)
}

func getQuery() string {
	query, exists := os.LookupEnv("QUERY")
	if !exists {
		query = "great vector database"
	}
	return query
}

func printResults(results []engine.ScoredDocument) {
	fmt.Printf("Scored documents: %d\n", len(results))
	fmt.Println(strings.Repeat("-", 22))
	fmt.Printf("| %-8s | %-8s |\n", "DocID", "Score")
	fmt.Println(strings.Repeat("-", 22))
	for _, doc := range results {
		fmt.Printf("| %-8d | %8.2f |\n", doc.DocID, doc.Score)
	}
	fmt.Println(strings.Repeat("-", 22))
}
