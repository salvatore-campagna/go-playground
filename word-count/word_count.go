// Package wordcount provides functionality for counting the frequency of words
// in a given phrase, while handling punctuation and special characters.
package wordcount

import (
	"strings"
	"unicode"
)

// Frequency represents a map where keys are words and values are their occurrence counts.
type Frequency map[string]int

// WordCount takes a phrase as input, converts it to lowercase, and counts the frequency
// of each word. Words are separated by spaces or punctuation, except for internal single quotes.
// Returns a Frequency map with words and their respective counts.
func WordCount(phrase string) Frequency {
	builder := strings.Builder{}
	wordCounts := make(Frequency)

	for _, r := range strings.ToLower(phrase) {
		if (unicode.IsSpace(r) || unicode.IsPunct(r)) && r != '\'' {
			updateWordFrequency(builder.String(), wordCounts)
			builder.Reset()
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' {
			// we need to keep the '\'' chracters here since we still don't know if they are
			// in the middle of a word or not. They will be removed later if they are
			// leading or trailing.
			builder.WriteRune(r)
		}
	}

	if builder.Len() > 0 {
		updateWordFrequency(builder.String(), wordCounts)
	}

	return wordCounts
}

// updateWordFrequency updates the count of a cleaned word in the given map.
// If the word is non-empty, it increments its count in the map.
func updateWordFrequency(input string, wordCounts map[string]int) {
	word := cleanWord(input)
	if word == "" {
		return
	}
	wordCounts[word]++
}

// cleanWord removes leading and trailing single quotes from a word while preserving
// those within the word. It returns the cleaned word.
func cleanWord(input string) string {
	return strings.Trim(input, "'")
}
