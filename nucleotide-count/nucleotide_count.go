// Package dna provides functionality to analyze DNA sequences and count nucleotide occurrences.
package dna

import (
	"fmt"
)

// Histogram is a mapping from a nucleotide rune ('A', 'C', 'G', 'T') to its count in a given DNA sequence.
type Histogram map[rune]int

// DNA represents a sequence of nucleotides ('A', 'C', 'G', 'T').
type DNA string

// Counts generates a histogram of valid nucleotides in the given DNA sequence.
// It returns an error if the DNA sequence contains any invalid characters.
// The method iterates over each character in the DNA string, counting occurrences of 'A', 'C', 'G', and 'T'.
func (dna DNA) Counts() (Histogram, error) {
	h := Histogram{
		'A': 0,
		'C': 0,
		'G': 0,
		'T': 0,
	}

	for _, r := range dna {
		switch r {
		case 'A', 'C', 'G', 'T':
			h[r]++
		default:
			return nil, fmt.Errorf("invalid DNA base %q", r)
		}
	}
	return h, nil
}
