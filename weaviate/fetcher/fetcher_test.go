package fetcher

import (
	"testing"
)

func TestParseJsonSegments(t *testing.T) {
	validJson := `{
		"segments": [
			[
				{
					"term": "vector",
					"doc_id": 1,
					"term_frequency": 0.5
				},
				{
					"term": "database",
					"doc_id": 2,
					"term_frequency": 0.7
				}
			],
			[
				{
					"term": "great",
					"doc_id": 1,
					"term_frequency": 0.3
				}
			]
		]
	}`

	segments, err := ParseTermPostings([]byte(validJson))
	if err != nil {
		t.Errorf("Failed to parse valid JSON: %v", err)
	}

	if len(segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(segments))
	}

	if len(segments[0]) != 2 {
		t.Errorf("Expected 2 term postings in first segment, got %d", len(segments[0]))
	}

	if len(segments[1]) != 1 {
		t.Errorf("Expected 1 term posting in second segment, got %d", len(segments[1]))
	}

	expectedTermPostings := []struct {
		segmentIndex int
		docIndex     int
		term         string
		docID        uint32
		termFreq     float32
	}{
		{0, 0, "vector", 1, 0.5},
		{0, 1, "database", 2, 0.7},
		{1, 0, "great", 1, 0.3},
	}

	for _, expectedTermPosting := range expectedTermPostings {
		doc := segments[expectedTermPosting.segmentIndex][expectedTermPosting.docIndex]
		if doc.Term != expectedTermPosting.term {
			t.Errorf("Expected term %s, got %s", expectedTermPosting.term, doc.Term)
		}
		if doc.DocID != expectedTermPosting.docID {
			t.Errorf("Expected docID %d, got %d", expectedTermPosting.docID, doc.DocID)
		}
		if doc.TermFrequency != expectedTermPosting.termFreq {
			t.Errorf("Expected frequency %f, got %f", expectedTermPosting.termFreq, doc.TermFrequency)
		}
	}
}

func TestEmptySegments(t *testing.T) {
	emptyJson := `{"segments":[]}`
	segments, err := ParseTermPostings([]byte(emptyJson))
	if err != nil {
		t.Errorf("Failed to parse empty segments: %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("Expected 0 segments, got %d", len(segments))
	}
}
