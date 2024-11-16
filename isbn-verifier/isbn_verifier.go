// Package isbn provides a function to validate ISBN-10 codes.
package isbn

import "unicode"

// IsValidISBN checks if a given string is a valid ISBN-10.
// An ISBN-10 is considered valid if it meets the following criteria:
// - It consists of exactly 10 valid characters, which can be digits (0-9) or 'X'.
// - Dashes ('-') are allowed but ignored during validation.
// - The character 'X' is only allowed as the last character and represents the value 10.
// - The checksum is calculated as: (10*d1 + 9*d2 + ... + 1*d10) % 11 == 0.
func IsValidISBN(isbn string) bool {
	checksum := 0
	digitIndex := 0
	isbnLen := len(isbn)

	for i, isbnDigit := range isbn {
		if isbnDigit == '-' {
			continue
		}

		if !isValidIsbnDigit(isbnDigit, i, isbnLen) {
			return false
		}

		checksum += digitValue(isbnDigit) * (10 - digitIndex)
		digitIndex++
	}

	return digitIndex == 10 && checksum%11 == 0
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

// digitValue converts a rune to its corresponding integer value.
// The character 'X' represents the value 10.
func digitValue(digit rune) int {
	if digit == 'X' {
		return 10
	}
	return int(digit - '0')
}
