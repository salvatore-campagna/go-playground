package storage

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"weaviate/fetcher"
)

func TestSegmentSerialization(t *testing.T) {
	// Generate random values for the RoaringBitmap
	values := generateRandomUint32Values(1000)

	// Create a new Segment
	segment := NewSegment()

	// Group documents by term before adding them
	termGroups := make(map[string][]fetcher.TermPosting)

	// First, group all documents by their term
	for value, included := range values {
		if included {
			doc := fetcher.TermPosting{
				Term:          fmt.Sprintf("term_%d", value),
				DocID:         value,
				TermFrequency: rand.Float32(),
			}
			termGroups[doc.Term] = append(termGroups[doc.Term], doc)
		}
	}

	// Then add each group of documents to the segment
	for _, docs := range termGroups {
		if len(docs) > 0 {
			segment.BulkIndex(docs)
		}
	}

	// Serialize the segment to a buffer
	var buffer bytes.Buffer
	if err := segment.WriteSegment(&buffer); err != nil {
		t.Fatalf("Failed to serialize segment: %v", err)
	}

	// Create a new segment and deserialize it from the buffer
	deserializedSegment := NewSegment()
	if err := deserializedSegment.ReadSegment(&buffer); err != nil {
		t.Fatalf("Failed to deserialize segment: %v", err)
	}

	// Verify that the term count matches after deserialization
	if len(deserializedSegment.Terms) != len(segment.Terms) {
		t.Errorf("Term count mismatch: got %d, expected %d",
			len(deserializedSegment.Terms), len(segment.Terms))
	}

	// Verify each term's data
	for term, termMetadata := range segment.Terms {
		deserializedTermMetadata, exists := deserializedSegment.Terms[term]
		if !exists {
			t.Errorf("Term %s not found in deserialized segment", term)
			continue
		}

		// Verify total docs count
		if termMetadata.TotalDocs != deserializedTermMetadata.TotalDocs {
			t.Errorf("TotalDocs mismatch for term %s: got %d, expected %d",
				term, deserializedTermMetadata.TotalDocs, termMetadata.TotalDocs)
		}

		// Verify blocks
		if len(termMetadata.Blocks) != len(deserializedTermMetadata.Blocks) {
			t.Errorf("Block count mismatch for term %s: got %d, expected %d",
				term, len(deserializedTermMetadata.Blocks), len(termMetadata.Blocks))
			continue
		}

		// Verify each block
		for i, block := range termMetadata.Blocks {
			deserializedBlock := deserializedTermMetadata.Blocks[i]

			if block.Bitmap == nil || deserializedBlock.Bitmap == nil {
				t.Errorf("Nil BitmapContainer found in block %d for term %s", i, term)
				continue
			}

			if block.Bitmap.Cardinality() != deserializedBlock.Bitmap.Cardinality() {
				t.Errorf("Cardinality mismatch in block %d for term %s: got %d, expected %d",
					i, term, deserializedBlock.Bitmap.Cardinality(), block.Bitmap.Cardinality())
			}
		}
	}
}

func TestSegmentTermLookup(t *testing.T) {
	segment := NewSegment()

	terms := []fetcher.TermPosting{
		{Term: "jedi", DocID: 123, TermFrequency: 0.5},
	}

	if err := segment.BulkIndex(terms); err != nil {
		t.Fatalf("failed to index documents: %v", err)
	}

	termIterator, err := segment.TermIterator("jedi")
	if err != nil {
		t.Fatalf("failed getting iterator for 'jedi': %v", err)
	}

	expectedDocuments := []struct {
		DocID         uint32
		TermFrequency float32
	}{
		{123, 0.5},
	}

	for _, expectedDocument := range expectedDocuments {
		hasNext, err := termIterator.Next()
		if err != nil {
			t.Fatalf("unexpected error while iterating segment: %v", err)
		}
		if !hasNext {
			t.Fatalf("unexpected missing value while iterating segment")
		}

		docID, err := termIterator.DocID()
		if err != nil {
			t.Fatalf("unexpected error while retrieving document ID for 'jedi': %v", err)
		}
		if docID != expectedDocument.DocID {
			t.Fatalf("unexpected DocID for 'jedi': got %d, expected %d", docID, expectedDocument.DocID)
		}

		termFrequency, err := termIterator.TermFrequency()
		if err != nil {
			t.Fatalf("unexpected error while retrieving term frequency for 'jedi': %v", err)
		}
		if termFrequency != expectedDocument.TermFrequency {
			t.Fatalf("unexpected term frequency for 'jedi': got %.2f, expected %.2f", termFrequency, expectedDocument.TermFrequency)
		}
	}

	hasNext, err := termIterator.Next()
	if err != nil {
		t.Fatalf("unexpected error when advancing iterator: %v", err)
	}
	if hasNext {
		t.Fatalf("expected iterator to be exhausted, but it still has elements")
	}
}

func TestSegmentTotalDocsTracking(t *testing.T) {
	segment := NewSegment()

	postings := []fetcher.TermPosting{
		{Term: "database", DocID: 1, TermFrequency: 1.5},
		{Term: "search", DocID: 1, TermFrequency: 0.7},
		{Term: "database", DocID: 2, TermFrequency: 2.0},
	}

	if err := segment.BulkIndex(postings); err != nil {
		t.Fatalf("Failed to index terms: %v", err)
	}

	expectedTotalDocs := uint32(2)
	if segment.TotalDocs() != expectedTotalDocs {
		t.Errorf("Expected TotalDocs %d, got %d", expectedTotalDocs, segment.TotalDocs())
	}
}

func TestBlockOverflowHandling(t *testing.T) {
	segment := NewSegment()
	postings := make([]fetcher.TermPosting, MaxDcoumentsPerBlock+1)

	for i := 0; i <= MaxDcoumentsPerBlock; i++ {
		postings[i] = fetcher.TermPosting{Term: "overflow", DocID: uint32(i + 1), TermFrequency: 1.0}
	}

	if err := segment.BulkIndex(postings); err != nil {
		t.Fatalf("Failed to index terms: %v", err)
	}

	termMetadata := segment.Terms["overflow"]
	if len(termMetadata.Blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(termMetadata.Blocks))
	}

	if termMetadata.Blocks[0].Bitmap.Cardinality() != MaxDcoumentsPerBlock {
		t.Errorf("Expected first block to have %d entries, got %d", MaxDcoumentsPerBlock, termMetadata.Blocks[0].Bitmap.Cardinality())
	}

	if termMetadata.Blocks[1].Bitmap.Cardinality() != 1 {
		t.Errorf("Expected second block to have 1 entry, got %d", termMetadata.Blocks[1].Bitmap.Cardinality())
	}
}

func TestTermFrequenciesConsistency(t *testing.T) {
	block := NewBlock()

	docIDs := []uint32{10, 20, 30, 40}
	frequencies := []float32{0.5, 0.8, 1.2, 1.5}

	for i, docID := range docIDs {
		if err := block.AddDocument(docID, frequencies[i]); err != nil {
			t.Fatalf("Failed to add document: %v", err)
		}
	}

	if block.Bitmap.Cardinality() != len(block.TermFrequencies) {
		t.Errorf("Mismatch between Bitmap cardinality (%d) and TermFrequencies length (%d)", block.Bitmap.Cardinality(), len(block.TermFrequencies))
	}
}

func TestSegmentSerialization2(t *testing.T) {
	segment := NewSegment()

	postings := []fetcher.TermPosting{
		{Term: "vector", DocID: 1, TermFrequency: 1.0},
		{Term: "vector", DocID: 2, TermFrequency: 2.0},
		{Term: "database", DocID: 3, TermFrequency: 1.5},
	}

	if err := segment.BulkIndex(postings); err != nil {
		t.Fatalf("Failed to index terms: %v", err)
	}

	buffer := &bytes.Buffer{}
	if err := segment.Serialize(buffer); err != nil {
		t.Fatalf("Failed to serialize segment: %v", err)
	}

	deserializedSegment := NewSegment()
	if err := deserializedSegment.Deserialize(buffer); err != nil {
		t.Fatalf("Failed to deserialize segment: %v", err)
	}

	if deserializedSegment.TotalDocs() != segment.TotalDocs() {
		t.Errorf("Expected TotalDocs %d after deserialization, got %d", segment.TotalDocs(), deserializedSegment.TotalDocs())
	}
}
