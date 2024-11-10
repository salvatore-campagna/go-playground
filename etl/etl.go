/*
Package etl provides a solution to the ETL (Extract-Transform-Load) exercise.

The ETL exercise focuses on transforming legacy data formats. Given a map where the keys are point values
and the values are slices of uppercase letters, the goal is to transform it into a new format where each
letter (in lowercase) is associated with its corresponding point value.

For example, given the input:

	map[int][]string{
	    1: {"A", "E", "I", "O", "U"},
	    2: {"D", "G"},
	}

The function should return:

	map[string]int{
	    "a": 1, "e": 1, "i": 1, "o": 1, "u": 1,
	    "d": 2, "g": 2,
	}
*/
package etl

import "strings"

// Transform takes a map with integer keys and slices of strings as values,
// and converts it into a map where each string is a lowercase key with its associated integer value.
//
// Parameters:
//   - in: A map[int][]string where the key represents a score and the value is a list of uppercase letters.
//
// Returns:
//   - A map[string]int where each letter from the input is converted to lowercase
//     and associated with its original score.
//
// Example:
//
//	Transform(map[int][]string{1: {"A", "E"}, 2: {"D", "G"}})
//	Output: map[string]int{"a": 1, "e": 1, "d": 2, "g": 2}
func Transform(in map[int][]string) map[string]int {
	letterToPoints := make(map[string]int)

	for points, letters := range in {
		for _, letter := range letters {
			letterToPoints[strings.ToLower(letter)] = points
		}
	}

	return letterToPoints
}
