// Package scrabble provides utilities to calculate the score of a word
// in the game of Scrabble.
package scrabble

import "strings"

// Score computes the total Scrabble score for the given word.
// It converts each letter to uppercase and calculates its value
// based on standard Scrabble letter scoring.
func Score(word string) int {
	totalScore := 0
	for _, r := range strings.ToUpper(word) {
		totalScore += letterValue(r)
	}
	return totalScore
}

// letterValue returns the Scrabble score for a given letter.
// The scoring is based on the standard Scrabble rules:
// - 1 point: A, E, I, O, U, L, N, S, T, R
// - 2 points: D, G
// - 3 points: B, C, M, P
// - 4 points: F, H, V, W, Y
// - 5 points: K
// - 8 points: J, X
// - 10 points: Q, Z
func letterValue(r rune) int {
	switch r {
	case 'D', 'G':
		return 2
	case 'B', 'C', 'M', 'P':
		return 3
	case 'F', 'H', 'V', 'W', 'Y':
		return 4
	case 'K':
		return 5
	case 'J', 'X':
		return 8
	case 'Q', 'Z':
		return 10
	default:
		return 1
	}
}
