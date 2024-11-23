package encoders

import (
	"bytes"
	"math/rand"
	"testing"
)

// generateRandomUint16Values generates a slice of random uint16 values.
func generateRandomUint16Values(n int) []uint16 {
	values := make([]uint16, n)
	for i := 0; i < n; i++ {
		values[i] = uint16(rand.Intn(65536))
	}
	return values
}

// generateMonotonicUint16Values generates a slice of monotonically increasing uint16 values
// with a random step between stepMin and stepMax.
func generateMonotonicUint16Values(n int, start, stepMin, stepMax uint16) []uint16 {
	values := make([]uint16, n)
	current := start
	for i := 0; i < n; i++ {
		values[i] = current
		step := uint16(rand.Intn(int(stepMax-stepMin+1))) + stepMin
		current += step
		if current > 65535 {
			current = 65535
		}
	}
	return values
}

// valuesAreEqual checks if two slices of uint16 are equal.
func valuesAreEqual(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDeltaEncoder tests the DeltaEncoder for both encoding and decoding with random values.
func TestDeltaEncoder_Serialization(t *testing.T) {
	original := generateRandomUint16Values(100)
	encoder := NewDeltaEncoder(0)

	var buffer bytes.Buffer
	if err := encoder.Encode(original, &buffer); err != nil {
		t.Fatalf("DeltaEncoder failed to encode: %v", err)
	}

	decoded, err := encoder.Decode(&buffer, len(original))
	if err != nil {
		t.Fatalf("DeltaEncoder failed to decode: %v", err)
	}

	if !valuesAreEqual(original, decoded) {
		t.Fatalf("DeltaEncoder failed: original and decoded values do not match")
	}
}

// TestPlainEncoder tests the PlainEncoder for both encoding and decoding with random values.
func TestPlainEncoder_Serialization(t *testing.T) {
	original := generateRandomUint16Values(100)
	encoder := NewPlainEncoder()

	var buffer bytes.Buffer
	if err := encoder.Encode(original, &buffer); err != nil {
		t.Fatalf("PlainEncoder failed to encode: %v", err)
	}

	decoded, err := encoder.Decode(&buffer, len(original))
	if err != nil {
		t.Fatalf("PlainEncoder failed to decode: %v", err)
	}

	if !valuesAreEqual(original, decoded) {
		t.Fatalf("PlainEncoder failed: original and decoded values do not match")
	}
}

// TestDeltaCompressionEfficiency ensures DeltaEncoder is more efficient than PlainEncoder for monotonic values.
func TestDeltaCompressionEfficiency(t *testing.T) {
	values := generateMonotonicUint16Values(5000, 0, 1, 10)

	var deltaBuffer, plainBuffer bytes.Buffer
	deltaEncoder := NewDeltaEncoder(0)
	plainEncoder := NewPlainEncoder()

	if err := deltaEncoder.Encode(values, &deltaBuffer); err != nil {
		t.Fatalf("DeltaEncoder failed to encode: %v", err)
	}
	if err := plainEncoder.Encode(values, &plainBuffer); err != nil {
		t.Fatalf("PlainEncoder failed to encode: %v", err)
	}

	deltaSize := deltaBuffer.Len()
	plainSize := plainBuffer.Len()
	t.Logf("Delta size: %d, Plain size: %d, Difference: %d", deltaSize, plainSize, plainSize-deltaSize)

	if deltaSize > plainSize {
		t.Fatalf("DeltaEncoder should produce fewer bytes than PlainEncoder for monotonic values")
	}
}

// TestDeltaEncoder_Fallback ensures DeltaEncoder falls back to PlainEncoder for short arrays.
func TestDeltaEncoder_Fallback(t *testing.T) {
	length := rand.Intn(10) + 1
	values := generateMonotonicUint16Values(length, 0, 1, 5)

	deltaEncoder := NewDeltaEncoder(length)
	var deltaBuffer, plainBuffer bytes.Buffer

	if err := deltaEncoder.Encode(values, &deltaBuffer); err != nil {
		t.Fatalf("DeltaEncoder failed to encode: %v", err)
	}
	plainEncoder := NewPlainEncoder()
	if err := plainEncoder.Encode(values, &plainBuffer); err != nil {
		t.Fatalf("PlainEncoder failed to encode: %v", err)
	}

	if !bytes.Equal(deltaBuffer.Bytes(), plainBuffer.Bytes()) {
		t.Fatalf("DeltaEncoder fallback failed: encoded bytes do not match PlainEncoder")
	}
}

// TestDeltaEncoder_EqualValues tests DeltaEncoder with identical values.
func TestDeltaEncoder_EqualValues(t *testing.T) {
	values := make([]uint16, 100)
	for i := range values {
		values[i] = 42
	}

	encoder := NewDeltaEncoder(0)
	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("DeltaEncoder failed to encode: %v", err)
	}

	decoded, err := encoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("DeltaEncoder failed to decode: %v", err)
	}

	if !valuesAreEqual(values, decoded) {
		t.Fatalf("DeltaEncoder failed for equal values")
	}
}

// TestDeltaEncoder_MonotonicValues tests DeltaEncoder with monotonically increasing values.
func TestDeltaEncoder_MonotonicValues(t *testing.T) {
	values := generateMonotonicUint16Values(100, 0, 1, 10)

	encoder := NewDeltaEncoder(0)
	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("DeltaEncoder failed to encode: %v", err)
	}

	decoded, err := encoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("DeltaEncoder failed to decode: %v", err)
	}

	if !valuesAreEqual(values, decoded) {
		t.Fatalf("DeltaEncoder failed for monotonic values")
	}
}

// TestPlainEncoder_PreserveIntegrity ensures input and output match for PlainEncoder.
func TestPlainEncoder_PreserveIntegrity(t *testing.T) {
	values := generateRandomUint16Values(100)
	encoder := NewPlainEncoder()

	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("PlainEncoder failed to encode: %v", err)
	}

	decoded, err := encoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("PlainEncoder failed to decode: %v", err)
	}

	if !valuesAreEqual(values, decoded) {
		t.Fatalf("PlainEncoder integrity check failed: input and output do not match")
	}
}

// TestDeltaEncoder_DecreasingValues tests DeltaEncoder with monotonically decreasing values.
func TestDeltaEncoder_DecreasingValues(t *testing.T) {
	values := make([]uint16, 100)
	for i := range values {
		values[i] = uint16(100 - i)
	}

	encoder := NewDeltaEncoder(0)
	var buffer bytes.Buffer
	if err := encoder.Encode(values, &buffer); err != nil {
		t.Fatalf("DeltaEncoder failed to encode: %v", err)
	}

	decoded, err := encoder.Decode(&buffer, len(values))
	if err != nil {
		t.Fatalf("DeltaEncoder failed to decode: %v", err)
	}

	if !valuesAreEqual(values, decoded) {
		t.Fatalf("DeltaEncoder failed for decreasing values")
	}
}
