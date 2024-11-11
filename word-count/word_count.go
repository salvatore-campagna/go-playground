// Package wordcount provides functionality for counting the frequency of words
// in a given phrase, handling punctuation and special characters.
package wordcount

import (
	"strings"
	"unicode"
)

// Frequency represents a map where keys are words and values are their occurrence counts.
type Frequency map[string]int

// WordCount takes a phrase as input, normalizes it to lowercase, and counts the frequency
// of each word. Words are separated by spaces or punctuation, with special handling for
// internal single quotes. Single quotes within words are preserved, while leading or
// trailing single quotes are removed during processing.
// Returns a Frequency map with words and their respective counts.
func WordCount(phrase string) Frequency {
	builder := strings.Builder{}
	frequency := make(Frequency)

	for _, r := range strings.ToLower(phrase) {
		if (unicode.IsSpace(r) || unicode.IsPunct(r)) && r != '\'' {
			updateWordFrequency(builder.String(), frequency)
			builder.Reset()
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' {
			// Preserve internal single quotes since they may be part of a word.
			// Leading or trailing quotes will be removed later.
			builder.WriteRune(r)
		}
	}

	if builder.Len() > 0 {
		updateWordFrequency(builder.String(), frequency)
	}

	return frequency
}

// updateWordFrequency updates the count of a word in the given frequency map.
// It trims leading and trailing single quotes from the word before counting.
// If the resulting word is empty, it is ignored.
func updateWordFrequency(input string, frequency Frequency) {
	word := strings.Trim(input, "'")
	if word == "" {
		return
	}
	frequency[word]++
}
