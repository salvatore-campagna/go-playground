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
const DefaultQuery = "great vector"

func main() {
	dir := flag.String("dir", DefaultSegmentDir, "Directory to load segment files from")
	query := flag.String("query", "", "Query terms (space-separated)")
	flag.Parse()

	effectiveQuery := *query
	if effectiveQuery == "" {
		effectiveQuery = DefaultQuery
	}

	files, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", *dir, err)
		return
	}

	var segments []*storage.Segment
	totalDocsBitmap := storage.NewRoaringBitmap()
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

		totalDocsBitmap = totalDocsBitmap.Union(segment.DocIDs)
		segments = append(segments, segment)
	}

	if len(segments) == 0 {
		fmt.Println("No valid segments found.")
		return
	}

	totalDocs := totalDocsBitmap.Cardinality()
	fmt.Printf("Total number of documents: %d\n", totalDocsBitmap.Cardinality())
	queryEngine, err := engine.NewQueryEngine(segments, uint32(totalDocs))
	if err != nil {
		panic(err)
	}

	terms := strings.Fields(effectiveQuery)

	fmt.Printf("Query: %s\n", effectiveQuery)
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
