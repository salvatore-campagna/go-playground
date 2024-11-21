package storage

// TODO: Add benchmarks for critical methods (e.g., Add, Serialize, Deserialize, Union, Intersection).
// Benchmarking can help identify bottlenecks and optimize performance-critical code.

import (
	"bytes"
	"math/bits"
	"math/rand"
	"testing"
)

const maxUint16 = 1 << 16

// Helper function to populate an ArrayContainer with values.
func populateArrayContainer(ac *ArrayContainer, values map[uint16]bool) {
	for value, included := range values {
		if included {
			ac.Add(value)
		}
	}
}

// Helper function to populate a BitmapContainer with values.
func populateBitmapContainer(bc *BitmapContainer, values map[uint16]bool) {
	for value, included := range values {
		if included {
			bc.Add(value)
		}
	}
}

// Helper function to populate a RoaringBitmap with values.
func populateRoaringBitmap(rb *RoaringBitmap, values map[uint32]bool) {
	for value, included := range values {
		if included {
			rb.Add(value)
		}
	}
}

// Helper function to generate random uint16 values for ArrayContainer and BitmapContainer.
func generateRandomUint16Values(max int) map[uint16]bool {
	values := make(map[uint16]bool)
	for len(values) < rand.Intn(max) {
		value := uint16(rand.Uint32() & 0xFFFF)
		values[value] = rand.Intn(2) == 0
	}
	return values
}

// Helper function to generate random uint32 values for RoaringBitmap.
func generateRandomUint32Values(max int) map[uint32]bool {
	values := make(map[uint32]bool)
	for len(values) < rand.Intn(max) {
		value := rand.Uint32()
		values[value] = rand.Intn(2) == 0
	}
	return values
}

//
// ArrayContainer Tests
//

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

	count := 0
	for value, included := range values {
		if included {
			ac.Add(value)
			count++
		}
	}

	if ac.Cardinality() != count {
		t.Errorf("Expected cardinality %d, got %d", count, ac.Cardinality())
	}
}

//
// BitmapContainer Tests
//

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

	count := 0
	for value, included := range values {
		if included {
			bc.Add(value)
			count++
		}
	}

	if bc.Cardinality() != count {
		t.Errorf("Expected cardinality %d, got %d", count, bc.Cardinality())
	}
}

//
// RoaringBitmap Tests
//

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

func TestRoaringBitmap_Contains(t *testing.T) {
	rb := NewRoaringBitmap()
	values := generateRandomUint32Values(10_000)

	for value, included := range values {
		if included {
			rb.Add(value)
		}
	}

	for value, included := range values {
		if included && !rb.Contains(value) {
			t.Errorf("RoaringBitmap does not contain expected value [%d]", value)
		}
		if !included && rb.Contains(value) {
			t.Errorf("RoaringBitmap incorrectly contains value [%d]", value)
		}
	}
}

func TestRoaringBitmap_Cardinality(t *testing.T) {
	rb := NewRoaringBitmap()
	values := generateRandomUint32Values(10_000)

	count := 0
	for value, included := range values {
		if included {
			rb.Add(value)
			count++
		}
	}

	if rb.Cardinality() != count {
		t.Errorf("Expected cardinality %d, got %d", count, rb.Cardinality())
	}
}

func TestRoaringBitmap_Serialization(t *testing.T) {
	originalRoaringBitmap := NewRoaringBitmap()
	values := generateRandomUint32Values(1000)
	populateRoaringBitmap(originalRoaringBitmap, values)

	var buffer bytes.Buffer
	if err := originalRoaringBitmap.Serialize(&buffer); err != nil {
		t.Fatalf("Failed to serialize RoaringBitmap: %v", err)
	}

	newRoaringBitmap := NewRoaringBitmap()
	if err := newRoaringBitmap.Deserialize(&buffer); err != nil {
		t.Fatalf("Failed to deserialize RoaringBitmap: %v", err)
	}

	if newRoaringBitmap.Cardinality() != originalRoaringBitmap.Cardinality() {
		t.Errorf("Cardinality mismatch after deserialization: got %d, expected %d",
			newRoaringBitmap.Cardinality(), originalRoaringBitmap.Cardinality())
	}

	for value, included := range values {
		if included && !newRoaringBitmap.Contains(value) {
			t.Errorf("Deserialized RoaringBitmap is missing value %d", value)
		} else if !included && newRoaringBitmap.Contains(value) {
			t.Errorf("Deserialized RoaringBitmap incorrectly includes value %d", value)
		}
	}
}

func TestArrayContainer_Rank(t *testing.T) {
	ac := &ArrayContainer{values: []uint16{1, 5, 10, 20, 30}}

	tests := []struct {
		input  uint16
		expect int
	}{
		{0, 0},
		{1, 1},
		{5, 2},
		{10, 3},
		{15, 3},
		{20, 4},
		{30, 5},
		{35, 5},
	}

	for _, test := range tests {
		result := ac.Rank(test.input)
		if result != test.expect {
			t.Errorf("ArrayContainer.Rank(%d) = %d; expected %d", test.input, result, test.expect)
		}
	}
}

func TestBitmapContainerRank(t *testing.T) {
	// Create a new BitmapContainer
	bc := NewBitmapContainer()

	// Add values to the bitmap
	bc.Add(1)  // Bit 1
	bc.Add(16) // Bit 16
	bc.Add(66) // Bit 66

	// Define test cases
	tests := []struct {
		input  uint16
		expect int
	}{
		{0, 0},
		{1, 1},
		{15, 1},
		{16, 2},
		{65, 2},
		{66, 3},
		{67, 3},
		{68, 3},
	}

	// Execute tests
	for _, test := range tests {
		result := bc.Rank(test.input)
		if result != test.expect {
			t.Errorf("BitmapContainer.Rank(%d) = %d; expected %d", test.input, result, test.expect)
		}
	}
}

func TestRoaringBitmap_Rank(t *testing.T) {
	rb := NewRoaringBitmap()

	rb.Add(0x00010001) // Key 0x0001, Low 0x0001
	rb.Add(0x00010002) // Key 0x0001, Low 0x0002
	rb.Add(0x00020003) // Key 0x0002, Low 0x0003
	rb.Add(0x00030004) // Key 0x0003, Low 0x0004

	tests := []struct {
		input  uint32
		expect int
	}{
		{0x00010001, 1},
		{0x00010002, 2},
		{0x00020003, 3},
		{0x00030004, 4},
		{0x00040000, 4},
	}

	for _, test := range tests {
		result, err := rb.Rank(test.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != test.expect {
			t.Errorf("RoaringBitmap.Rank(%#x) = %d; expected %d", test.input, result, test.expect)
		}
	}
}

func TestBitmapContainerRankRandomSelection(t *testing.T) {
	bc := NewBitmapContainer()

	// Populate the bitmap with alternating bits (101010...)
	for i := uint16(0); i < 65535; i++ {
		if i%2 == 0 {
			bc.Add(i)
		}
	}

	for i := 0; i < 1_000; i++ {
		targetLow := uint16(rand.Intn(65536))
		expectedRank := calculateExpectedRank(bc.Bitmap, targetLow)
		result := bc.Rank(targetLow)
		if result != expectedRank {
			t.Errorf("BitmapContainer.Rank(%d) = %d; expected %d", targetLow, result, expectedRank)
		}
	}
}

// Helper function to calculate the expected rank
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
