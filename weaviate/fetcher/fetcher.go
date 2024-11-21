package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
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
	// Check if the path is a URL (starts with "http" or "https")
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

	// Treat it as a local file path
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
