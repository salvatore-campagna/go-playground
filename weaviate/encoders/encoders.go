// Package encoders provides implementations for encoding and decoding arrays of uint16 values
// using different compression techniques like Delta and Plain encoding. These encoders are useful
// for efficiently storing sequences of numbers, particularly in scenarios where data can benefit
// from delta compression (e.g., sorted or sequential data).
package encoders

// TODO: experiment with other encoders

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// ArrayEncoder defines the interface for encoding an array of uint16 values to a writer.
type ArrayEncoder interface {
	// Encode encodes the given array of uint16 values and writes it to the provided writer.
	// It returns an error if any encoding or writing operation fails.
	Encode(values []uint16, writer io.Writer) error
}

// ArrayDecoder defines the interface for decoding an array of uint16 values from a reader.
type ArrayDecoder interface {
	// Decode reads a specified number of uint16 values from the reader and returns them as an array.
	// It returns an error if any reading or decoding operation fails.
	Decode(reader io.Reader, length int) ([]uint16, error)
}

// ArrayEncoderDecoder combines both encoding and decoding methods into one interface.
type ArrayEncoderDecoder interface {
	ArrayEncoder
	ArrayDecoder
}

// DeltaEncoder implements ArrayEncoder and ArrayDecoder using delta encoding with varint compression.
// Delta encoding is a compression technique where each value in the sequence is stored as the difference
// from the previous value, which is then encoded using variable-length integers.
type DeltaEncoder struct {
	minLen          int
	fallbackEncoder ArrayEncoderDecoder
}

// NewDeltaEncoder creates and returns a new instance of DeltaEncoder.
func NewDeltaEncoder(minLen int) *DeltaEncoder {
	return &DeltaEncoder{
		minLen:          minLen,
		fallbackEncoder: NewPlainEncoder(),
	}
}

// Encode compresses the given array of uint16 values using delta encoding and varint encoding.
// The first value is stored as-is, while subsequent values are stored as the difference from the previous value.
func (d *DeltaEncoder) Encode(values []uint16, writer io.Writer) error {
	if len(values) <= d.minLen {
		fmt.Printf("Fallback")
		return d.fallbackEncoder.Encode(values, writer)
	}

	if err := binary.Write(writer, binary.LittleEndian, values[0]); err != nil {
		return err
	}

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

// Decode reads a delta-varint encoded array of uint16 values from the reader and reconstructs the original values.
func (d *DeltaEncoder) Decode(reader io.Reader, length int) ([]uint16, error) {
	if length == 0 {
		return []uint16{}, nil
	}

	values := make([]uint16, length)

	// Read the first value as-is
	if err := binary.Read(reader, binary.LittleEndian, &values[0]); err != nil {
		return nil, err
	}

	// Decode deltas using varint
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

// writeVarint writes a uint64 value using varint encoding.
func writeVarint(writer io.Writer, value uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	_, err := writer.Write(buf[:n])
	return err
}

// readVarint reads a uint64 value using varint decoding.
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

// PlainEncoder implements ArrayEncoder and ArrayDecoder using plain encoding.
// Plain encoding writes the values as they are without any compression.
type PlainEncoder struct{}

// NewPlainEncoder creates and returns a new instance of PlainEncoder.
func NewPlainEncoder() *PlainEncoder {
	return &PlainEncoder{}
}

// Encode writes the given array of uint16 values directly to the writer without any compression.
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
