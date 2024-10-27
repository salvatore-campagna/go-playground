// Package hamming provides functionality to calculate the Hamming distance
// between two DNA strands. The Hamming distance is the number of differing
// characters at corresponding positions in two DNA strands of equal length.
package hamming

import "fmt"

// Distance calculates the Hamming distance between two equal-length DNA strands.
// It returns an error if the DNA strands are of different lengths.
//
// The Hamming distance is defined as the number of positions where the
// corresponding characters differ between the two DNA strands.
//
// Example:
//
//	dist, err := Distance("GAGCCT", "CATCGT")
//	// dist will be 4, err will be nil
//	dist, err := Distance("GAG", "CATC")
//	// dist will be -1, err will be not nill
func Distance(a, b string) (int, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("DNA strands must have the same length: %d and %d provided", len(a), len(b))
	}

	distance := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			distance++
		}
	}

	return distance, nil
}
