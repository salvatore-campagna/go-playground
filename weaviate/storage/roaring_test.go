package storage

import (
	"bytes"
	"math/bits"
	"math/rand"
	"sort"
	"testing"
)

const (
	maxUint16 = ^uint16(0)
	maxUint32 = ^uint32(0)
)

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

// calculateExpectedBitmapRank calculates the expected rank for a BitmapContainer.
func calculateExpectedBitmapRank(bitmap []uint64, value uint16) int {
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

// calculateExpectedArrayRank calculates the expected rank for an ArrayContainer.
func calculateExpectedArrayRank(values []uint16, value uint16) int {
	return sort.Search(len(values), func(i int) bool { return values[i] > value })
}

// ArrayContainer Tests

func TestArrayContainer_AddContaines(t *testing.T) {
	ac := NewArrayContainer()
	values := generateRandomUint16Values(10_000)

	populateArrayContainer(ac, values)

	for value, included := range values {
		if !included && ac.Contains(value) {
			t.Errorf("Array container includes unexpected value %d", value)
		}
		if included && !ac.Contains(value) {
			t.Errorf("Array container missing value %d", value)
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

func TestArrayContainer_Rank(t *testing.T) {
	ac := NewArrayContainer()

	for i := uint16(0); i < uint16(2*1024); i++ {
		if i%2 == 0 {
			ac.Add(i)
		}
	}

	for i := 0; i < 1_000; i++ {
		value := uint16(rand.Intn(1 << 16))
		expectedRank := calculateExpectedArrayRank(ac.values, value)
		actualRank := ac.Rank(value)
		if expectedRank != actualRank {
			t.Errorf("Array container rank(%d) = %d, expected %d", value, actualRank, expectedRank)
		}
	}
}

func TestArrayContainer_Union(t *testing.T) {
	bc1 := NewArrayContainer()
	bc2 := NewArrayContainer()
	values := generateRandomUint16Values(1_000)

	expectedCount := 0
	unionValues := make(map[uint16]bool)

	for value, included := range values {
		if included {
			addToBC1 := rand.Intn(2) == 0
			addToBC2 := rand.Intn(2) == 0

			if addToBC1 {
				bc1.Add(value)
				unionValues[value] = true
			}
			if addToBC2 {
				bc2.Add(value)
				unionValues[value] = true
			}

			if addToBC1 || addToBC2 {
				expectedCount++
			}
		}
	}

	union := bc1.Union(bc2)
	for value := range unionValues {
		if !union.Contains(value) {
			t.Errorf("BitmapContainer union missing value %d", value)
		}
	}

	if union.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, union.Cardinality())
	}
}

func TestArrayContainer_Intersection(t *testing.T) {
	ac1 := NewArrayContainer()
	ac2 := NewArrayContainer()
	values := generateRandomUint16Values(1_000)

	expectedCount := 0
	intersectingValues := make(map[uint16]bool)

	for value, included := range values {
		if included {
			addToAC1 := rand.Intn(2) == 0
			addToAC2 := rand.Intn(2) == 0

			if addToAC1 {
				ac1.Add(value)
			}
			if addToAC2 {
				ac2.Add(value)
			}

			if addToAC1 && addToAC2 {
				intersectingValues[value] = true
				expectedCount++
			}
		}
	}

	intersection := ac1.Intersection(ac2)
	for value, included := range intersectingValues {
		if included && !intersection.Contains(value) {
			t.Errorf("Array container intersection missing value %d", value)
		}
	}

	if intersection.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, intersection.Cardinality())
	}
}

// BitmapContainer Tests

func TestBitmapContainer_AddContains(t *testing.T) {
	bc := NewBitmapContainer()
	values := generateRandomUint16Values(10_000)

	populateBitmapContainer(bc, values)

	for value, included := range values {
		if included && !bc.Contains(value) {
			t.Errorf("Bitmap container does not contain expected value %d", value)
		}
		if !included && bc.Contains(value) {
			t.Errorf("Bitmap container incorrectly contains value %d", value)
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

	for i := uint16(0); i < uint16(16*1024-1); i++ {
		if i%2 == 0 {
			bc.Add(i)
		}
	}

	for i := 0; i < 1_000; i++ {
		value := uint16(rand.Intn(1 << 16))
		expectedRank := calculateExpectedBitmapRank(bc.Bitmap, value)
		actualRank := bc.Rank(value)
		if expectedRank != actualRank {
			t.Errorf("Bitmap container rank(%d) = %d, expected %d", value, actualRank, expectedRank)
		}
	}
}

func TestBitmapContainer_Union(t *testing.T) {
	bc1 := NewBitmapContainer()
	bc2 := NewBitmapContainer()
	values := generateRandomUint16Values(1_000)

	expectedCount := 0
	unionValues := make(map[uint16]bool)

	for value, included := range values {
		if included {
			addToBC1 := rand.Intn(2) == 0
			addToBC2 := rand.Intn(2) == 0

			if addToBC1 {
				bc1.Add(value)
				unionValues[value] = true
			}
			if addToBC2 {
				bc2.Add(value)
				unionValues[value] = true
			}

			if addToBC1 || addToBC2 {
				expectedCount++
			}
		}
	}

	union := bc1.Union(bc2)
	for value := range unionValues {
		if !union.Contains(value) {
			t.Errorf("Bitmap container union missing value %d", value)
		}
	}

	if union.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, union.Cardinality())
	}
}

func TestBitmapContainer_Intersection(t *testing.T) {
	bc1 := NewBitmapContainer()
	bc2 := NewBitmapContainer()
	values := generateRandomUint16Values(1_000)

	expectedCount := 0
	intersectionValues := make(map[uint16]bool)

	for value, included := range values {
		if included {
			addToBC1 := rand.Intn(2) == 0
			addToBC2 := rand.Intn(2) == 0

			if addToBC1 {
				bc1.Add(value)
			}
			if addToBC2 {
				bc2.Add(value)
			}

			if addToBC1 && addToBC2 {
				intersectionValues[value] = true
				expectedCount++
			}
		}
	}

	intersection := bc1.Intersection(bc2)
	for value := range intersectionValues {
		if !intersection.Contains(value) {
			t.Errorf("Bitmap container intersection missing value %d", value)
		}
	}

	if intersection.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, intersection.Cardinality())
	}
}

// RoaringBitmap Tests

func TestRoaringBitmap_Empty(t *testing.T) {
	rb := NewRoaringBitmap()

	if rb.Cardinality() != 0 {
		t.Errorf("Empty bitmap should have cardinality 0, got %d", rb.Cardinality())
	}

	if rb.Contains(0) {
		t.Errorf("Empty bitmap should not contain any value, but contains 0")
	}
}

func TestRoaringBitmap_BoundaryValues(t *testing.T) {
	rb := NewRoaringBitmap()
	values := []uint32{
		0, uint32(maxUint16) - 1, uint32(maxUint16), uint32(maxUint16) + 1, uint32(maxUint32),
	}

	for _, value := range values {
		rb.Add(value)
	}

	for _, value := range values {
		if !rb.Contains(value) {
			t.Errorf("Bitmap should contain %d", value)
		}
	}

	if rb.Cardinality() != len(values) {
		t.Errorf("Expected cardinality %d, got %d", len(values), rb.Cardinality())
	}
}

func TestRoaringBitmap_SerializationEmpty(t *testing.T) {
	original := NewRoaringBitmap()

	var buffer bytes.Buffer
	if err := original.Serialize(&buffer); err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	deserialized := NewRoaringBitmap()
	if err := deserialized.Deserialize(&buffer); err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if deserialized.Cardinality() != 0 {
		t.Errorf("Deserialized empty bitmap should have cardinality 0, got %d", deserialized.Cardinality())
	}
}

func TestRoaringBitmap_ContainerTransition(t *testing.T) {
	rb := NewRoaringBitmap()

	for i := uint32(0); i < 4097; i++ {
		rb.Add(i)
	}

	if rb.Cardinality() != 4097 {
		t.Errorf("Expected cardinality 4097, got %d", rb.Cardinality())
	}

	for i := uint32(0); i < 4097; i++ {
		if !rb.Contains(i) {
			t.Errorf("Bitmap missing value %d", i)
		}
	}
}

func TestRoaringBitmap_Add(t *testing.T) {
	rb := NewRoaringBitmap()
	values := generateRandomUint32Values(10_000)

	populateRoaringBitmap(rb, values)

	for value, included := range values {
		if included && !rb.Contains(value) {
			t.Errorf("Roaring bitmap missing value %d", value)
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
			t.Errorf("Deserialized bitmap missing value %d", value)
		}
	}
}

func TestRoaringBitmap_Union(t *testing.T) {
	rb1 := NewRoaringBitmap()
	rb2 := NewRoaringBitmap()
	values := generateRandomUint32Values(1_000)

	expectedCount := 0
	unionValues := make(map[uint32]bool)

	for value, included := range values {
		if included {
			addToRB1 := rand.Intn(2) == 0
			addToRB2 := rand.Intn(2) == 0

			if addToRB1 {
				rb1.Add(value)
				unionValues[value] = true
			}
			if addToRB2 {
				rb2.Add(value)
				unionValues[value] = true
			}

			if addToRB1 || addToRB2 {
				expectedCount++
			}
		}
	}

	union := rb1.Union(rb2)
	for value := range unionValues {
		if !union.Contains(value) {
			t.Errorf("Roaring bitmap union missing value %d", value)
		}
	}

	if union.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, union.Cardinality())
	}
}

func TestRoaringBitmap_Intersection(t *testing.T) {
	rb1 := NewRoaringBitmap()
	rb2 := NewRoaringBitmap()
	values := generateRandomUint32Values(1_000)

	expectedCount := 0
	intersectionValues := make(map[uint32]bool)

	for value, included := range values {
		if included {
			addToRB1 := rand.Intn(2) == 0
			addToRB2 := rand.Intn(2) == 0

			if addToRB1 {
				rb1.Add(value)
			}
			if addToRB2 {
				rb2.Add(value)
			}

			if addToRB1 && addToRB2 {
				intersectionValues[value] = true
				expectedCount++
			}
		}
	}

	intersection := rb1.Intersection(rb2)
	for value := range intersectionValues {
		if !intersection.Contains(value) {
			t.Errorf("Roaring bitmap intersection missing value %d", value)
		}
	}

	if intersection.Cardinality() != expectedCount {
		t.Errorf("Expected cardinality %d, got %d", expectedCount, intersection.Cardinality())
	}
}
