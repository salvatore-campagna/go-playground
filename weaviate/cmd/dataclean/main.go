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
func ParseJsonSegments(data []byte) (Root, error) {
	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		return root, fmt.Errorf("failed to parse json: %w", err)
	}
	return root, nil
}

// CleanSegments removes duplicate document IDs from the segments
func CleanSegments(root Root) Root {
	uniqueDocIDs := make(map[uint32]struct{})
	cleanedSegments := make([][]JsonDocument, len(root.Segments))

	for i, segment := range root.Segments {
		uniqueDocs := []JsonDocument{}
		for _, doc := range segment {
			if _, exists := uniqueDocIDs[doc.DocID]; !exists {
				uniqueDocIDs[doc.DocID] = struct{}{}
				uniqueDocs = append(uniqueDocs, doc)
			}
		}
		cleanedSegments[i] = uniqueDocs
	}

	return Root{Segments: cleanedSegments}
}

// WriteJsonToFile writes the cleaned segments to a JSON file
func WriteJsonToFile(root Root, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(root); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func main() {
	inputFilePath := flag.String("input", "", "Path to the input JSON file")
	outputFilePath := flag.String("output", "", "Path to the output JSON file")
	flag.Parse()

	if *inputFilePath == "" || *outputFilePath == "" {
		log.Fatalf("Both input and output file paths must be specified")
	}

	data, err := FetchJson(*inputFilePath)
	if err != nil {
		log.Fatalf("Error fetching JSON: %v", err)
	}

	root, err := ParseJsonSegments(data)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	cleanedRoot := CleanSegments(root)

	if err := WriteJsonToFile(cleanedRoot, *outputFilePath); err != nil {
		log.Fatalf("Error writing cleaned JSON to file: %v", err)
	}

	fmt.Printf("Cleaned JSON file written successfully to: %s\n", *outputFilePath)
}
