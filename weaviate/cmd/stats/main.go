package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// JsonDocument represents a single entry in the segment JSON
type JsonDocument struct {
	Term          string  `json:"term"`
	DocID         uint32  `json:"doc_id"`
	TermFrequency float32 `json:"term_frequency"`
}

// Root represents the top-level structure of the JSON file
type Root struct {
	Segments [][]JsonDocument `json:"segments"`
}

// Statistics to hold computed stats
type Statistics struct {
	TotalSegments          int
	TotalDocuments         map[uint32]struct{}
	TotalRepeatedDocuments map[uint32]struct{}
	TotalTerms             map[string]struct{}
	DocFrequencyPerTerm    map[string]int
	DocumentsPerSegment    []map[uint32]struct{}
	TermsPerSegment        []map[string]struct{}
	DocFrequencyPerSegment []map[string]int
}

// FetchJson fetches JSON data from either a URL or a local file path.
func FetchJson(path string) ([]byte, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		response, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json: %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-ok HTTP response: %s", response.Status)
		}

		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return data, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read local file: %w", err)
	}
	return data, nil
}

// ParseJsonSegments parses the JSON data into a slice of segments
func ParseJsonSegments(data []byte) ([][]JsonDocument, error) {
	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}
	return root.Segments, nil
}

// ComputeStatistics calculates the required statistics from the segments
func ComputeStatistics(segments [][]JsonDocument) Statistics {
	stats := Statistics{
		TotalSegments:          len(segments),
		TotalDocuments:         make(map[uint32]struct{}),
		TotalRepeatedDocuments: make(map[uint32]struct{}),
		TotalTerms:             make(map[string]struct{}),
		DocFrequencyPerTerm:    make(map[string]int),
		DocumentsPerSegment:    make([]map[uint32]struct{}, len(segments)),
		TermsPerSegment:        make([]map[string]struct{}, len(segments)),
		DocFrequencyPerSegment: make([]map[string]int, len(segments)),
	}

	for i, segment := range segments {
		stats.DocumentsPerSegment[i] = make(map[uint32]struct{})
		stats.TermsPerSegment[i] = make(map[string]struct{})
		stats.DocFrequencyPerSegment[i] = make(map[string]int)

		for _, doc := range segment {
			if _, exists := stats.TotalDocuments[doc.DocID]; exists {
				stats.TotalRepeatedDocuments[doc.DocID] = struct{}{}
				continue
			}
			stats.TotalDocuments[doc.DocID] = struct{}{}
			stats.TotalTerms[doc.Term] = struct{}{}
			stats.DocumentsPerSegment[i][doc.DocID] = struct{}{}
			stats.TermsPerSegment[i][doc.Term] = struct{}{}
			stats.DocFrequencyPerTerm[doc.Term]++
			stats.DocFrequencyPerSegment[i][doc.Term]++
		}
	}

	return stats
}

func main() {
	inputFilePath := flag.String("path", "", "Path to the input JSON file")
	flag.Parse()

	if *inputFilePath == "" {
		log.Fatalf("Input file path must be specified using the -path flag")
	}

	data, err := FetchJson(*inputFilePath)
	if err != nil {
		log.Fatalf("Error fetching JSON: %v", err)
	}

	segments, err := ParseJsonSegments(data)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	stats := ComputeStatistics(segments)

	// Print the computed statistics in a nice tabular format
	fmt.Printf("\n+============== Stats ===============\n\n")

	fmt.Printf("Total Segments: %d\n\n", stats.TotalSegments)

	fmt.Printf("Segment\tDistinct Docs\tDistinct Terms\n")
	fmt.Printf("-------\t-------------\t--------------\n")
	for i := 0; i < stats.TotalSegments; i++ {
		fmt.Printf("%d\t%d\t\t%d\n", i, len(stats.DocumentsPerSegment[i]), len(stats.TermsPerSegment[i]))
	}

	fmt.Printf("\nTotal Documents: %d\n", len(stats.TotalDocuments))
	fmt.Printf("Total Repeated Documents: %d\n", len(stats.TotalRepeatedDocuments))
	fmt.Printf("Total Terms: %d\n\n", len(stats.TotalTerms))

	format := fmt.Sprintf("%%-15s\t%%-15d\n")
	fmt.Printf("%-15s\t%-15s\n", "Term", "Doc Frequency")
	fmt.Printf("-------------\t-------------\n")
	for term, freq := range stats.DocFrequencyPerTerm {
		fmt.Printf(format, term, freq)
	}
}
