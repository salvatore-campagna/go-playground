package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// TermPosting represents a single entry in the segment JSON.
type TermPosting struct {
	Term          string  `json:"term"`
	DocID         uint32  `json:"doc_id"`
	TermFrequency float32 `json:"term_frequency"`
}

// TermPostingRoot represents the top-level structure of the JSON file.
type TermPostingRoot struct {
	Segments [][]TermPosting `json:"segments"`
}

// FetchJson fetches JSON data from a URL or a local file.
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

// ParseTermPostings parses JSON data into a slice of term posting segments.
func ParseTermPostings(data []byte) ([][]TermPosting, error) {
	var root TermPostingRoot
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}
	return root.Segments, nil
}
