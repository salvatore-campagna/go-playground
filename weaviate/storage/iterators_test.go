package storage

import (
	"math/rand"
	"sort"
	"testing"
)

func TestEmptyRoaringBitmapIterator(t *testing.T) {
	it := &RoaringBitmapIterator{
		bitmap: NewRoaringBitmap(),
		term:   "test",
	}

	hasNext, err := it.Next()
	if err != nil {
		t.Fatalf("Unexpected error during iteration: %v", err)
	}

	if hasNext {
		t.Errorf("Expected 'false' for empty iterator, but got: %v", hasNext)
	}
}

func TestBitmapIterator_BelowThreshold(t *testing.T) {
	bitmap := NewRoaringBitmap()

	for i := 0; i < 4096; i++ {
		bitmap.Add(uint32(i))
	}

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	for i := 0; i < 4096; i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}
		if docID != uint32(i) {
			t.Errorf("Expected DocID %d, but got %d", i, docID)
		}
	}

	// Ensure iterator is exhausted
	hasNext, err := it.Next()
	if hasNext || err != nil {
		t.Errorf("Expected iterator to be exhausted, but Next returned: hasNext=%v, err=%v", hasNext, err)
	}
}

func TestBitmapIterator_AboveThreshold(t *testing.T) {
	bitmap := NewRoaringBitmap()

	for i := 0; i < 8192; i++ {
		bitmap.Add(uint32(i))
	}

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	for i := 0; i < 8192; i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}
		if docID != uint32(i) {
			t.Errorf("Expected DocID %d, but got %d", i, docID)
		}
	}

	// Ensure iterator is exhausted
	hasNext, err := it.Next()
	if hasNext || err != nil {
		t.Errorf("Expected iterator to be exhausted, but Next returned: hasNext=%v, err=%v", hasNext, err)
	}
}

func TestBitmapIteratorRandom_MultipleContainers(t *testing.T) {
	bitmap := NewRoaringBitmap()
	expectedValues := make([]uint32, 0)

	for i := 0; i < 16*1024; i++ {
		value := rand.Uint32() & 0xFFFFF
		expectedValues = append(expectedValues, value)
		bitmap.Add(value)
	}

	// Ensure uniqueness and sort values (doc IDs in a container are in sort order)
	expectedValues = removeDuplicates(expectedValues)
	sort.Slice(expectedValues, func(i, j int) bool { return expectedValues[i] < expectedValues[j] })

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	for i := 0; i < len(expectedValues); i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}
		if docID != expectedValues[i] {
			t.Errorf("Expected DocID %d, but got %d", expectedValues[i], docID)
		}
	}

	// Ensure iterator is exhausted
	hasNext, err := it.Next()
	if hasNext || err != nil {
		t.Errorf("Expected iterator to be exhausted, but Next returned: hasNext=%v, err=%v", hasNext, err)
	}
}

func TestBitmapIteratorSparseValues(t *testing.T) {
	bitmap := NewRoaringBitmap()

	values := []uint32{1, 10, 100, 1000, 5000, 15000, 30000, 50000, 100_000, 1_000_000, 5_000_000, 1_000_000_000, 4_000_000_000}
	for _, value := range values {
		bitmap.Add(value)
	}

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	for i, expected := range values {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}
		if docID != expected {
			t.Errorf("Expected DocID %d, but got %d", expected, docID)
		}
	}

	// Ensure iterator is exhausted
	hasNext, err := it.Next()
	if hasNext || err != nil {
		t.Errorf("Expected iterator to be exhausted, but Next returned: hasNext=%v, err=%v", hasNext, err)
	}
}

func TestBitmapIteratorSingleValueContainers(t *testing.T) {
	bitmap := NewRoaringBitmap()

	expectedDocIDs := []uint32{}
	for i := 0; i < 10; i++ {
		docID := uint32(i * 65536)
		bitmap.Add(docID)
		expectedDocIDs = append(expectedDocIDs, docID)
	}

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	for i, expectedDocID := range expectedDocIDs {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}

		if docID != expectedDocID {
			t.Errorf("Expected DocID %d, but got %d", expectedDocID, docID)
		}
	}

	// Ensure iterator is exhausted
	hasNext, err := it.Next()
	if hasNext || err != nil {
		t.Errorf("Expected iterator to be exhausted, but Next returned: hasNext=%v, err=%v", hasNext, err)
	}
}

func TestBitmapIteratorSingleContainer(t *testing.T) {
	bitmap := NewRoaringBitmap()

	for i := 0; i < 100; i++ {
		bitmap.Add(uint32(i))
	}

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	for i := 0; i < 100; i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}
		if docID != uint32(i) {
			t.Errorf("Expected DocID %d, but got %d", i, docID)
		}
	}

	// Ensure iterator is exhausted
	hasNext, err := it.Next()
	if hasNext || err != nil {
		t.Errorf("Expected iterator to be exhausted, but Next returned: hasNext=%v, err=%v", hasNext, err)
	}
}

func TestBitmapIteratorTermFrequency(t *testing.T) {
	bitmap := NewRoaringBitmap()

	for i := 10_000; i < 16_000; i++ {
		bitmap.Add(uint32(i))
	}

	randomTermFrequency := rand.Float32()
	it := NewRoaringBitmapIterator(bitmap, "test", randomTermFrequency)
	for i := 10_000; i < 16_000; i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			t.Fatalf("Iterator terminated prematurely at index %d", i)
		}

		termFrequency, err := it.TermFrequency()
		if err != nil {
			t.Fatalf("Unexpected error retrieving TermFrequency: %v", err)
		}

		if termFrequency != randomTermFrequency {
			t.Errorf("Expected TermFrequency 3.0, but got %.2f", termFrequency)
		}
	}
}

func TestBitmapIteratorComplexContainers(t *testing.T) {
	bitmap := NewRoaringBitmap()
	for i := 0; i < 4000; i++ {
		bitmap.Add(uint32(i)) // Dense container
	}
	for i := 0; i < 1000; i += 10 {
		bitmap.Add(uint32(65536 + i)) // Sparse container
	}
	for i := 0; i < 10; i++ {
		bitmap.Add(uint32(131072 + i*1000)) // Very sparse container
	}

	it := NewRoaringBitmapIterator(bitmap, "test", 2.0)
	var docIDs []uint32
	for {
		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if !hasNext {
			break
		}

		docID, err := it.DocID()
		if err != nil {
			t.Fatalf("Unexpected error retrieving DocID: %v", err)
		}
		docIDs = append(docIDs, docID)
	}

	expectedDocIDs := []uint32{}
	for i := 0; i < 4000; i++ {
		expectedDocIDs = append(expectedDocIDs, uint32(i))
	}
	for i := 0; i < 1000; i += 10 {
		expectedDocIDs = append(expectedDocIDs, uint32(65536+i))
	}
	for i := 0; i < 10; i++ {
		expectedDocIDs = append(expectedDocIDs, uint32(131072+i*1000))
	}

	if len(docIDs) != len(expectedDocIDs) {
		t.Fatalf("Mismatch in number of DocIDs. Expected %d, got %d", len(expectedDocIDs), len(docIDs))
	}

	for i, docID := range docIDs {
		if docID != expectedDocIDs[i] {
			t.Errorf("Expected DocID %d at index %d, but got %d", expectedDocIDs[i], i, docID)
		}
	}
}

// Helper: Remove duplicates from a slice
func removeDuplicates(slice []uint32) []uint32 {
	unique := make(map[uint32]struct{})
	result := make([]uint32, 0)

	for _, value := range slice {
		if _, exists := unique[value]; !exists {
			unique[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}
