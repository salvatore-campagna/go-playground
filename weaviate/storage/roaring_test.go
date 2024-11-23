package storage

import (
	"bytes"
	"math/bits"
	"math/rand"
	"testing"
)

// Constants
const maxUint16 = 1 << 16

// Helper Functions

// populateArrayContainer populates an ArrayContainer with the given values.
func populateArrayContainer(ac *ArrayContainer, values map[uint16]bool) {
	for value, included := range values {
		if included {
			ac.Add(value)
		}
	}
}

// populateBitmapContainer populates a BitmapContainer with the given values.
func populateBitmapContainer(bc *BitmapContainer, values map[uint16]bool) {
	for value, included := range values {
		if included {
			bc.Add(value)
		}
	}
}

// populateRoaringBitmap populates a RoaringBitmap with the given values.
func populateRoaringBitmap(rb *RoaringBitmap, values map[uint32]bool) {
	for value, included := range values {
		if included {
			rb.Add(value)
		}
	}
}

// generateRandomUint16Values generates a map of random uint16 values.
func generateRandomUint16Values(max int) map[uint16]bool {
	values := make(map[uint16]bool)
	for len(values) < rand.Intn(max) {
		value := uint16(rand.Uint32() & 0xFFFF)
		values[value] = rand.Intn(2) == 0
	}
	return values
}

// generateRandomUint32Values generates a map of random uint32 values.
func generateRandomUint32Values(max int) map[uint32]bool {
	values := make(map[uint32]bool)
	for len(values) < rand.Intn(max) {
		value := rand.Uint32()
		values[value] = rand.Intn(2) == 0
	}
	return values
}

// calculateExpectedRank calculates the expected rank for a BitmapContainer.
func calculateExpectedRank(bitmap []uint64, value uint16) int {
	rank := 0
	for wordIndex := 0; wordIndex <= int(value/64); wordIndex++ {
		word := bitmap[wordIndex]
		if wordIndex == int(value/64) {
			mask := (uint64(1) << ((value % 64) + 1)) - 1
			word &= mask
		}
		rank += bits.OnesCount64(word)
	}
	return rank
}

// ArrayContainer Tests

func TestArrayContainer_Add(t *testing.T) {
	ac := NewArrayContainer()
	values := generateRandomUint16Values(10_000)

	populateArrayContainer(ac, values)

	for value, included := range values {
		if included && !ac.Contains(value) {
			t.Errorf("ArrayContainer missing value [%d] after adding", value)
		}
	}
}

func TestArrayContainer_Contains(t *testing.T) {
	ac := NewArrayContainer()
	values := generateRandomUint16Values(10_000)

	populateArrayContainer(ac, values)

	for value, included := range values {
		if included && !ac.Contains(value) {
			t.Errorf("ArrayContainer does not contain expected value [%d]", value)
		}
		if !included && ac.Contains(value) {
			t.Errorf("ArrayContainer incorrectly contains value [%d]", value)
		}
	}
}

func TestArrayContainer_Cardinality(t *testing.T) {
	ac := NewArrayContainer()
	values := generateRandomUint16Values(10_000)

	expectedCount := 0
	for value, included := range values {
		if included {
			ac.Add(value)
			expectedCount++
		}
	}

	if ac.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, ac.Cardinality())
	}
}

// BitmapContainer Tests

func TestBitmapContainer_Add(t *testing.T) {
	bc := NewBitmapContainer()
	values := generateRandomUint16Values(10_000)

	populateBitmapContainer(bc, values)

	for value, included := range values {
		if included && !bc.Contains(value) {
			t.Errorf("BitmapContainer missing value [%d] after adding", value)
		}
	}
}

func TestBitmapContainer_Contains(t *testing.T) {
	bc := NewBitmapContainer()
	values := generateRandomUint16Values(10_000)

	populateBitmapContainer(bc, values)

	for value, included := range values {
		if included && !bc.Contains(value) {
			t.Errorf("BitmapContainer does not contain expected value [%d]", value)
		}
		if !included && bc.Contains(value) {
			t.Errorf("BitmapContainer incorrectly contains value [%d]", value)
		}
	}
}

func TestBitmapContainer_Cardinality(t *testing.T) {
	bc := NewBitmapContainer()
	values := generateRandomUint16Values(10_000)

	expectedCount := 0
	for value, included := range values {
		if included {
			bc.Add(value)
			expectedCount++
		}
	}

	if bc.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, bc.Cardinality())
	}
}

func TestBitmapContainer_Rank(t *testing.T) {
	bc := NewBitmapContainer()

	// Add alternating bits
	for i := uint16(0); i < uint16(16*1024-1); i++ {
		if i%2 == 0 {
			bc.Add(i)
		}
	}

	// Validate rank calculation
	for i := 0; i < 1_000; i++ {
		value := uint16(rand.Intn(maxUint16))
		expectedRank := calculateExpectedRank(bc.Bitmap, value)
		actualRank := bc.Rank(value)
		if expectedRank != actualRank {
			t.Errorf("BitmapContainer.Rank(%d) = %d; expected %d", value, actualRank, expectedRank)
		}
	}
}

// RoaringBitmap Tests

func TestRoaringBitmap_Add(t *testing.T) {
	rb := NewRoaringBitmap()
	values := generateRandomUint32Values(10_000)

	populateRoaringBitmap(rb, values)

	for value, included := range values {
		if included && !rb.Contains(value) {
			t.Errorf("RoaringBitmap missing value [%d] after adding", value)
		}
	}
}

func TestRoaringBitmap_Serialization(t *testing.T) {
	rb := NewRoaringBitmap()
	values := generateRandomUint32Values(1_000)

	populateRoaringBitmap(rb, values)

	var buffer bytes.Buffer
	if err := rb.Serialize(&buffer); err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	deserializedRB := NewRoaringBitmap()
	if err := deserializedRB.Deserialize(&buffer); err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if rb.Cardinality() != deserializedRB.Cardinality() {
		t.Errorf("Cardinality mismatch: expected %d, got %d", rb.Cardinality(), deserializedRB.Cardinality())
	}

	for value, included := range values {
		if included && !deserializedRB.Contains(value) {
			t.Errorf("Deserialized bitmap missing value [%d]", value)
		}
	}
}
