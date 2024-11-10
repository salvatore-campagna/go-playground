/*
Package strand provides functions for DNA to RNA transcription.
*/
package strand

import "strings"

/*
ToRNA converts a DNA strand into its RNA complement.

Each nucleotide in the DNA strand is transcribed to its RNA complement:
  - 'G' -> 'C'
  - 'C' -> 'G'
  - 'T' -> 'A'
  - 'A' -> 'U'

Example:

	dna := "GCTA"
	rna := ToRNA(dna)
	// rna is "CGAU"

If an empty string is provided, it returns an empty string.
*/
func ToRNA(dna string) string {
	var sb strings.Builder
	sb.Grow(len(dna))

	for _, r := range dna {
		switch r {
		case 'G':
			sb.WriteRune('C')
		case 'C':
			sb.WriteRune('G')
		case 'T':
			sb.WriteRune('A')
		case 'A':
			sb.WriteRune('U')
		}
	}

	return sb.String()
}
