package storage

import (
	"math/rand"
	"sort"
	"testing"
)

func TestEmptyRoaringBitmapIterator(t *testing.T) {
	bitmap := NewRoaringBitmap()
	it := NewRoaringBitmapIterator(bitmap)

	hasNext, err := it.Next()
	if err != nil {
		t.Errorf("unexpected error while iterating bitmap")
	}

	if hasNext {
		t.Errorf("expected 'true' but ietrator next returns: %v", hasNext)
	}
}

func TestBitmapIteratorRandomInput_BelowThreashold(t *testing.T) {
	bitmap := NewRoaringBitmap()

	for i := 0; i < 4096; i++ {
		bitmap.Add(uint32(i))
	}

	it := NewRoaringBitmapIterator(bitmap)
	for i := 0; i < 4096; i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Errorf("unexpected error while iterating bitmap")
		}

		if !hasNext {
			t.Errorf("expected true but iterator returned: %v", hasNext)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Errorf("unexpected error while retriving DocID")
		}
		if docID != uint32(i) {
			t.Errorf("expected DocID %d, actual: %d", uint32(i), docID)
		}
	}
}

func TestBitmapIteratorRandomInput_AboveThreshold(t *testing.T) {
	bitmap := NewRoaringBitmap()

	for i := 0; i < 8192; i++ {
		bitmap.Add(uint32(i))
	}

	it := NewRoaringBitmapIterator(bitmap)
	for i := 0; i < 8192; i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Errorf("unexpected error while iterating bitmap")
		}

		if !hasNext {
			t.Errorf("expected true but iterator returned: %v", hasNext)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Errorf("unexpected error while retriving DocID")
		}
		if docID != uint32(i) {
			t.Errorf("expected DocID %d, actual: %d", uint32(i), docID)
		}
	}
}

func TestBitmapIterator_MultipleContainers(t *testing.T) {
	bitmap := NewRoaringBitmap()
	expectedValues := make([]uint32, 0)

	for i := 0; i < 16*1024; i++ {
		expectedValue := rand.Uint32()
		expectedValues = append(expectedValues, expectedValue)
		bitmap.Add(expectedValue)
	}

	// DocIDs are sorted and unique
	expectedValues = removeDuplicates(expectedValues)
	sort.Slice(expectedValues, func(i, j int) bool {
		return expectedValues[i] < expectedValues[j]
	})

	it := NewRoaringBitmapIterator(bitmap)
	for i := 0; i < len(expectedValues); i++ {
		hasNext, err := it.Next()
		if err != nil {
			t.Errorf("unexpected error while iterating bitmap")
		}

		if !hasNext {
			t.Errorf("expected true but iterator returned: %v", hasNext)
		}

		docID, err := it.DocID()
		if err != nil {
			t.Errorf("unexpected error while retriving DocID")
		}
		if docID != expectedValues[i] {
			t.Errorf("expected DocID %d, actual: %d", expectedValues[i], docID)
		}
	}
}

func removeDuplicates(slice []uint32) []uint32 {
	unique := make(map[uint32]bool)
	var result []uint32

	for _, value := range slice {
		if _, exists := unique[value]; !exists {
			unique[value] = true
			result = append(result, value)
		}
	}

	return result
}
