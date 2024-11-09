// Package isogram provides a function to check if a word is an isogram.
// An isogram is a word with no repeating letters, regardless of case,
// and ignoring spaces, hyphens, and other special characters.
package isogram

import (
	"strings"
	"unicode"
)

// IsIsogram determines if a given word is an isogram.
// An isogram is a word without repeating letters.
// Only letters are considered; special characters, spaces, and punctuation are ignored.
// The function returns true if the word is an isogram, and false otherwise.
func IsIsogram(word string) bool {
	runes := make(map[rune]bool, 26)

	for _, r := range strings.ToLower(word) {
		if !unicode.IsLetter(r) {
			continue
		}
		if _, exists := runes[r]; exists {
			return false
		}
		runes[r] = true
	}

	return true
}
