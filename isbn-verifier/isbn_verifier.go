// Package isbn provides a function to validate ISBN-10 codes.
package isbn

import "unicode"

// IsValidISBN checks if a given string is a valid ISBN-10.
// An ISBN-10 is considered valid if it meets the following criteria:
// - It consists of exactly 10 valid characters, which can be digits (0-9) or 'X'.
// - Dashes ('-') are allowed but ignored during validation.
// - The character 'X' is only allowed as the last character and represents the value 10.
// - The checksum is calculated as: (10*d1 + 9*d2 + ... + 1*d10) % 11 == 0.
// Returns true if the ISBN-10 is valid, otherwise returns false.
func IsValidISBN(isbn string) bool {
	value := 0
	digitIndex := 0
	isbnLen := len(isbn)

	for i, r := range isbn {
		if r == '-' {
			continue
		}

		if !isValidIsbnDigit(r, i, isbnLen) {
			return false
		}

		var digitValue int
		if r == 'X' {
			digitValue = 10
		} else {
			digitValue = int(r - '0')
		}

		value += digitValue * (10 - digitIndex)
		digitIndex++
	}

	return digitIndex == 10 && value%11 == 0
}

// isValidIsbnDigit checks if a character is a valid ISBN digit.
// The character 'X' is only valid if it appears as the last character.
// Otherwise, it must be a numeric digit (0-9).
func isValidIsbnDigit(digit rune, index, length int) bool {
	if digit == 'X' {
		return index == length-1
	}
	return unicode.IsDigit(digit)
}
