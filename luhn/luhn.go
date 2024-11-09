// Package luhn provides a function to validate strings using the Luhn algorithm.
// The Luhn algorithm is used to validate identification numbers such as credit card numbers.
package luhn

import (
	"strings"
	"unicode"
)

// Valid checks if the provided string is valid according to the Luhn algorithm.
// It ignores spaces and requires that the input consists of only digits (after removing spaces).
// The function returns true if the string is a valid Luhn sequence, otherwise false.
func Valid(id string) bool {
	_id := removeAllSpaces(id)
	if len(_id) <= 1 {
		return false
	}

	luhnValue := 0
	shouldDouble := len(_id)%2 == 0

	for _, r := range _id {
		if !unicode.IsDigit(r) {
			return false
		}

		digit := int(r - '0')
		if shouldDouble {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		luhnValue += digit
		shouldDouble = !shouldDouble
	}

	return luhnValue%10 == 0
}

// removeAllSpaces removes all whitespace characters (including Unicode spaces)
// from the provided string and returns the resulting string.
func removeAllSpaces(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
