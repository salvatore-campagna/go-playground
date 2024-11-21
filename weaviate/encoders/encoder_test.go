package encoders

import (
	"bytes"
	"math/rand"
	"testing"
)

// Helper function to generate a slice of random uint16 values
func generateRandomUint16Values(n int) []uint16 {
	values := make([]uint16, n)
	for i := 0; i < n; i++ {
		values[i] = uint16(rand.Intn(65536))
	}
	return values
}

// Helper function to generate a slice of monotonically increasing uint16 values
// with a random step between stepMin and stepMax.
func generateMonotonicUint16Values(n int, start uint16, stepMin uint16, stepMax uint16) []uint16 {
	values := make([]uint16, n)
	currentValue := start

	for i := 0; i < n; i++ {
		values[i] = currentValue
		step := uint16(rand.Intn(int(stepMax-stepMin+1))) + stepMin
		currentValue += step

		// Ensure we wrap around if we exceed the max value of uint16
		if currentValue > 65535 {
			currentValue = 65535
		}
	}

	return values
}

// Helper function to check if two slices are equal
func valuesAreEqual(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDeltaEncoder tests the DeltaEncoder for both encoding and decoding
func TestDeltaEncoder_Serialization(t *testing.T) {
	originalValues := generateRandomUint16Values(100)
	encoder := NewDeltaEncoder(0)

	var buffer bytes.Buffer
	if err := encoder.Encode(originalValues, &buffer); err != nil {
		t.Fatalf("Failed to serialize with DeltaEncoder: %v", err)
	}

	decoder := NewDeltaEncoder(0)
	decodedValues, err := decoder.Decode(&buffer, len(originalValues))
	if err != nil {
		t.Fatalf("Failed to deserialize with DeltaEncoder: %v", err)
	}

	if !valuesAreEqual(originalValues, decodedValues) {
		t.Fatalf("DeltaEncoder serialization/deserialization failed: original and decoded values do not match.")
	}
}

// TestPlainEncoder tests the PlainEncoder for both encoding and decoding
func TestPlainEncoder_Serialization(t *testing.T) {
	originalValues := generateRandomUint16Values(100)
	encoder := NewPlainEncoder()

	var buffer bytes.Buffer
	if err := encoder.Encode(originalValues, &buffer); err != nil {
		t.Fatalf("Failed to serialize with PlainEncoder: %v", err)
	}

	decoder := NewPlainEncoder()
	decodedValues, err := decoder.Decode(&buffer, len(originalValues))
	if err != nil {
		t.Fatalf("Failed to deserialize with PlainEncoder: %v", err)
	}

	if !valuesAreEqual(originalValues, decodedValues) {
		t.Fatalf("PlainEncoder serialization/deserialization failed: original and decoded values do not match.")
	}
}

// TestDeltaCompressionEfficiency checks if delta compression results in fewer bytes than plain encoding
func TestDeltaCompressionEfficiency(t *testing.T) {
	values := generateMonotonicUint16Values(1000, 0, 0, 10)

	// Encode with DeltaEncoder
	deltaEncoder := NewDeltaEncoder(0)
	var deltaBuffer bytes.Buffer
	if err := deltaEncoder.Encode(values, &deltaBuffer); err != nil {
		t.Fatalf("Failed to serialize with DeltaEncoder: %v", err)
	}

	// Encode with PlainEncoder
	plainEncoder := NewPlainEncoder()
	var plainBuffer bytes.Buffer
	if err := plainEncoder.Encode(values, &plainBuffer); err != nil {
		t.Fatalf("Failed to serialize with PlainEncoder: %v", err)
	}

	deltaSize := deltaBuffer.Len()
	plainSize := plainBuffer.Len()
	diff := plainSize - deltaSize

	t.Logf("Delta Buffer Length: %d, Plain Buffer Length: %d, Difference: %d bytes", deltaSize, plainSize, diff)

	if deltaSize >= plainSize {
		t.Fatalf("DeltaEncoder should produce fewer or equal number of bytes than PlainEncoder, but got Delta: %d, Plain: %d", deltaSize, plainSize)
	}
}

// TestDeltaCompressionEfficiency checks if delta compression results in fewer bytes than plain encoding
func TestNoDeltaEncodingBelowMinLen(t *testing.T) {
	len := 1 + rand.Intn(128)
	values := generateMonotonicUint16Values(len, 0, 0, 10)

	// Encode with DeltaEncoder
	deltaEncoder := NewDeltaEncoder(len)
	var deltaBuffer bytes.Buffer
	if err := deltaEncoder.Encode(values, &deltaBuffer); err != nil {
		t.Fatalf("Failed to serialize with DeltaEncoder: %v", err)
	}

	// Encode with PlainEncoder
	plainEncoder := NewPlainEncoder()
	var plainBuffer bytes.Buffer
	if err := plainEncoder.Encode(values, &plainBuffer); err != nil {
		t.Fatalf("Failed to serialize with PlainEncoder: %v", err)
	}

	if !bytes.Equal(deltaBuffer.Bytes(), plainBuffer.Bytes()) {
		t.Fatalf("unexpected delta encoding")
	}
}

// TestDeltaEncoder_EqualValues tests delta encoding with a series of equal values
func TestDeltaEncoder_EqualValues(t *testing.T) {
	values := make([]uint16, 100)
	for i := range values {
		values[i] = 42
	}

	encoder := NewDeltaEncoder(0)
	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("Failed to serialize with DeltaEncoder: %v", err)
	}

	decoder := NewDeltaEncoder(0)
	decodedValues, err := decoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("Failed to deserialize with DeltaEncoder: %v", err)
	}

	if !valuesAreEqual(values, decodedValues) {
		t.Fatalf("DeltaEncoder failed with equal values: original and decoded values do not match.")
	}
}

// TestDeltaEncoder_IncreasingValues tests delta encoding with monotonically increasing values
func TestDeltaEncoder_IncreasingValues(t *testing.T) {
	values := make([]uint16, 100)
	for i := range values {
		values[i] = uint16(i)
	}

	encoder := NewDeltaEncoder(0)
	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("Failed to serialize with DeltaEncoder: %v", err)
	}

	decoder := NewDeltaEncoder(0)
	decodedValues, err := decoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("Failed to deserialize with DeltaEncoder: %v", err)
	}

	if !valuesAreEqual(values, decodedValues) {
		t.Fatalf("DeltaEncoder failed with increasing values: original and decoded values do not match.")
	}
}

// TestDeltaEncoder_DecreasingValues tests delta encoding with monotonically decreasing values
func TestDeltaEncoder_DecreasingValues(t *testing.T) {
	values := make([]uint16, 100)
	for i := range values {
		values[i] = uint16(100 - i)
	}

	encoder := NewDeltaEncoder(0)
	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("Failed to serialize with DeltaEncoder: %v", err)
	}

	decoder := NewDeltaEncoder(0)
	decodedValues, err := decoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("Failed to deserialize with DeltaEncoder: %v", err)
	}

	if !valuesAreEqual(values, decodedValues) {
		t.Fatalf("DeltaEncoder failed with decreasing values: original and decoded values do not match.")
	}
}

// TestPlainEncoder_PreserveIntegrity checks that input and output are identical for PlainEncoder
func TestPlainEncoder_PreserveIntegrity(t *testing.T) {
	values := generateRandomUint16Values(100)
	encoder := NewPlainEncoder()

	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("Failed to serialize with PlainEncoder: %v", err)
	}

	decoder := NewPlainEncoder()
	decodedValues, err := decoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("Failed to deserialize with PlainEncoder: %v", err)
	}

	if !valuesAreEqual(values, decodedValues) {
		t.Fatalf("PlainEncoder failed to preserve integrity: original and decoded values do not match.")
	}
}
