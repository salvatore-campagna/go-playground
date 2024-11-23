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

func TestBitmapIteratorSequential_BelowThreshold(t *testing.T) {
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

func TestBitmapIteratorSequential_AboveThreshold(t *testing.T) {
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

	values := []uint32{100, 1000, 5000, 15000, 30000, 50000}
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
