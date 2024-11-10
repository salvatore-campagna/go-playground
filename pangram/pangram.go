/*
Package pangram provides a function to determine if a given string is a pangram.

A pangram is a sentence that contains every letter of the English alphabet at least once.
For example:
- "The quick brown fox jumps over the lazy dog" is a pangram.
- "Hello, world!" is not a pangram because it does not contain all the letters of the alphabet.
*/
package pangram

import (
	"strings"
	"unicode"
)

// IsPangram checks if the given input string is a pangram.
//
// A pangram is defined as a sentence that includes every letter of the English alphabet
// at least once. This function ignores case, spaces, and non-letter characters.
//
// Parameters:
//   - input: The string to be checked for being a pangram.
//
// Returns:
//   - bool: true if the input string is a pangram, false otherwise.
//
// Example usage:
//
//	pangram.IsPangram("The quick brown fox jumps over the lazy dog") // Returns true
//	pangram.IsPangram("Hello, world!")                               // Returns false
func IsPangram(input string) bool {
	var characters [26]bool
	charactersCount := 0

	for _, r := range strings.ToLower(input) {
		if unicode.IsLetter(r) {
			index := r - 'a'
			if !characters[index] {
				characters[index] = true
				charactersCount++

				if charactersCount == 26 {
					return true
				}
			}
		}
	}

	return charactersCount == 26
}
