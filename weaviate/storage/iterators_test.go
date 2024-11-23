package storage

import (
	"math/rand"
	"sort"
	"testing"
)

func TestEmptyRoaringBitmapIterator(t *testing.T) {
	t.Run("Empty Iterator", func(t *testing.T) {
		bitmap := NewRoaringBitmap()
		it := NewRoaringBitmapIterator(bitmap)

		hasNext, err := it.Next()
		if err != nil {
			t.Fatalf("Unexpected error during iteration: %v", err)
		}

		if hasNext {
			t.Errorf("Expected 'false' for empty iterator, but got: %v", hasNext)
		}
	})
}

func TestBitmapIteratorSequential_BelowThreshold(t *testing.T) {
	t.Run("Sequential Input Below Threshold", func(t *testing.T) {
		bitmap := NewRoaringBitmap()

		for i := 0; i < 4096; i++ {
			bitmap.Add(uint32(i))
		}

		it := NewRoaringBitmapIterator(bitmap)
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
	})
}

func TestBitmapIteratorSequential_AboveThreshold(t *testing.T) {
	t.Run("Sequential Input Above Threshold", func(t *testing.T) {
		bitmap := NewRoaringBitmap()

		for i := 0; i < 8192; i++ {
			bitmap.Add(uint32(i))
		}

		it := NewRoaringBitmapIterator(bitmap)
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
	})
}

func TestBitmapIteratorRandom_MultipleContainers(t *testing.T) {
	t.Run("Random Input Across Multiple Containers", func(t *testing.T) {
		bitmap := NewRoaringBitmap()
		expectedValues := make([]uint32, 0)

		for i := 0; i < 16*1024; i++ {
			value := rand.Uint32() & 0xFFFFF // Mask to limit to 20-bit range
			expectedValues = append(expectedValues, value)
			bitmap.Add(value)
		}

		// Ensure uniqueness and sort values
		expectedValues = removeDuplicates(expectedValues)
		sort.Slice(expectedValues, func(i, j int) bool { return expectedValues[i] < expectedValues[j] })

		it := NewRoaringBitmapIterator(bitmap)
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
	})
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
