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

	// Check the number of segments
	if len(segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(segments))
	}

	// Check the number of documents in each segment
	if len(segments[0]) != 2 {
		t.Errorf("Expected 2 documents in first segment, got %d", len(segments[0]))
	}

	if len(segments[1]) != 1 {
		t.Errorf("Expected 1 document in second segment, got %d", len(segments[1]))
	}

	// Check the contents of the documents
	expectedDocs := []struct {
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

	for _, expected := range expectedDocs {
		doc := segments[expected.segmentIndex][expected.docIndex]
		if doc.Term != expected.term {
			t.Errorf("Expected term %s, got %s", expected.term, doc.Term)
		}
		if doc.DocID != expected.docID {
			t.Errorf("Expected docID %d, got %d", expected.docID, doc.DocID)
		}
		if doc.TermFrequency != expected.termFreq {
			t.Errorf("Expected frequency %f, got %f", expected.termFreq, doc.TermFrequency)
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
