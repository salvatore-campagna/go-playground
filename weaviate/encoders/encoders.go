// Package encoders provides implementations for encoding and decoding arrays of uint16 values
// using different compression techniques such as Delta and Plain encoding. These encoders are
// designed for efficient storage and transmission of sequences of numbers, particularly when
// data exhibits patterns that benefit from compression (e.g., sorted or sequential values).
//
// # Encoding Techniques
//
//  1. **Delta Encoding**: Stores the difference between consecutive values. This is efficient
//     for sequences with small or predictable increments. Combined with variable-length encoding
//     (varint), it achieves further compression.
//
//  2. **Plain Encoding**: Writes values directly without compression. Suitable for datasets
//     where values are random or do not exhibit compressible patterns.
//
// # TODOs
//
//   - Explore additional encoding techniques such as Run-Length Encoding (RLE)
//     for datasets with repeated values.
//   - Evaluate the performance and space efficiency of hybrid encoders that
//     dynamically choose between Delta and Plain encoding based on input characteristics.
//   - Investigate SIMD (Single Instruction, Multiple Data) optimizations for faster encoding and decoding.
//   - Add support for other data types (e.g., uint32, float32) for broader applicability.
//   - Benchmark encoding and decoding implementations under various scenarios
//     to determine performance bottlenecks and optimize accordingly.
//   - Implement error handling improvements, including more descriptive error messages
//     for boundary conditions like overflow in varint encoding.
package encoders

import (
	"encoding/binary"
	"errors"
	"io"
)

// ArrayEncoder defines the interface for encoding an array of uint16 values to a writer.
type ArrayEncoder interface {
	// Encode encodes the given array of uint16 values and writes it to the provided writer.
	// Returns an error if encoding or writing fails.
	Encode(values []uint16, writer io.Writer) error
}

// ArrayDecoder defines the interface for decoding an array of uint16 values from a reader.
type ArrayDecoder interface {
	// Decode reads a specified number of uint16 values from the reader and reconstructs the array.
	// Returns an error if decoding or reading fails.
	Decode(reader io.Reader, length int) ([]uint16, error)
}

// ArrayEncoderDecoder combines both encoding and decoding methods into one interface.
type ArrayEncoderDecoder interface {
	ArrayEncoder
	ArrayDecoder
}

// DeltaEncoder implements ArrayEncoder and ArrayDecoder using delta encoding
// combined with varint compression. It compresses data by storing differences
// between consecutive values instead of the values themselves.
type DeltaEncoder struct {
	minLen          int
	fallbackEncoder ArrayEncoderDecoder
}

// NewDeltaEncoder creates and returns a new DeltaEncoder. If the input array length
// is below `minLen`, it falls back to plain encoding.
func NewDeltaEncoder(minLen int) *DeltaEncoder {
	return &DeltaEncoder{
		minLen:          minLen,
		fallbackEncoder: NewPlainEncoder(),
	}
}

// Encode compresses the given array of uint16 values using delta encoding and writes it
// to the provided writer. If the array length is below `minLen`, it falls back to plain encoding.
func (d *DeltaEncoder) Encode(values []uint16, writer io.Writer) error {
	if len(values) <= d.minLen {
		return d.fallbackEncoder.Encode(values, writer)
	}

	// Write the first value directly
	if err := binary.Write(writer, binary.LittleEndian, values[0]); err != nil {
		return err
	}

	// Write subsequent values as deltas
	prev := values[0]
	for i := 1; i < len(values); i++ {
		delta := values[i] - prev
		prev = values[i]
		if err := writeVarint(writer, uint64(delta)); err != nil {
			return err
		}
	}
	return nil
}

// Decode reads a delta-encoded array of uint16 values and reconstructs the original array.
// The first value is read directly, and subsequent values are reconstructed using deltas.
func (d *DeltaEncoder) Decode(reader io.Reader, length int) ([]uint16, error) {
	if length == 0 {
		return []uint16{}, nil
	}

	values := make([]uint16, length)

	// Read the first value
	if err := binary.Read(reader, binary.LittleEndian, &values[0]); err != nil {
		return nil, err
	}

	// Decode deltas and reconstruct the array
	prev := values[0]
	for i := 1; i < length; i++ {
		delta, err := readVarint(reader)
		if err != nil {
			return nil, err
		}
		values[i] = prev + uint16(delta)
		prev = values[i]
	}
	return values, nil
}

// PlainEncoder implements ArrayEncoder and ArrayDecoder by writing and reading
// uint16 values without any compression.
type PlainEncoder struct{}

// NewPlainEncoder creates and returns a new PlainEncoder.
func NewPlainEncoder() *PlainEncoder {
	return &PlainEncoder{}
}

// Encode writes the given array of uint16 values directly to the writer without compression.
func (p *PlainEncoder) Encode(values []uint16, writer io.Writer) error {
	for _, v := range values {
		if err := binary.Write(writer, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	return nil
}

// Decode reads a specified number of uint16 values from the reader and returns them as an array.
func (p *PlainEncoder) Decode(reader io.Reader, length int) ([]uint16, error) {
	values := make([]uint16, length)
	for i := 0; i < length; i++ {
		if err := binary.Read(reader, binary.LittleEndian, &values[i]); err != nil {
			return nil, err
		}
	}
	return values, nil
}

// writeVarint writes a uint64 value using varint encoding. Varint encodes integers
// into a variable number of bytes, optimizing storage for smaller numbers.
func writeVarint(writer io.Writer, value uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	_, err := writer.Write(buf[:n])
	return err
}

// readVarint reads a uint64 value encoded using varint from the reader. It decodes
// the variable-length representation back into the original integer value.
func readVarint(reader io.Reader) (uint64, error) {
	var value uint64
	var buf [1]byte
	shift := uint(0)

	for {
		if _, err := reader.Read(buf[:]); err != nil {
			return 0, err
		}
		b := buf[0]
		value |= (uint64(b&0x7F) << shift)
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if shift > 64 {
			return 0, errors.New("varint overflow")
		}
	}
	return value, nil
}
