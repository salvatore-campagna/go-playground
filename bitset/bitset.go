/*
Package bitset provides a simple implementation of a bitset using a slice of uint64 integers.

A bitset is a memory-efficient data structure for storing bits (0 or 1) and is useful for representing sets of integers.
*/
package bitset

import (
	"fmt"
	"math/bits"
)

// BitSet represents a bitset using a slice of uint64 values.
type BitSet struct {
	bits []uint64
}

// NewBitSet creates a BitSet with the given size (in bits).
func NewBitSet(size int) *BitSet {
	return &BitSet{
		bits: make([]uint64, (size+63)/64),
	}
}

// Set sets the bit at the specified position to 1.
func (bs *BitSet) Set(pos int) error {
	index, offset := pos/64, pos%64
	if index < 0 || index >= len(bs.bits) {
		return fmt.Errorf("invalid position: %d", pos)
	}
	bs.bits[index] |= 1 << offset
	return nil
}

// Clear resets the bit at the specified position to 0.
func (bs *BitSet) Clear(pos int) error {
	index, offset := pos/64, pos%64
	if index < 0 || index >= len(bs.bits) {
		return fmt.Errorf("invalid position: %d", pos)
	}
	bs.bits[index] &^= 1 << offset
	return nil
}

// Test returns true if the bit at the specified position is set to 1.
func (bs *BitSet) Test(pos int) (bool, error) {
	index, offset := pos/64, pos%64
	if index < 0 || index >= len(bs.bits) {
		return false, fmt.Errorf("invalid position: %d", pos)
	}
	return (bs.bits[index] & (1 << offset)) != 0, nil
}

// Count returns the number of bits set to 1.
func (bs *BitSet) Count() int {
	count := 0
	for _, word := range bs.bits {
		count += bits.OnesCount64(word)
	}
	return count
}
