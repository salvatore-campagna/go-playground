/*
Package romannumerals provides a function to convert integers into their corresponding Roman numeral representations.

Roman numerals use letters from the Latin alphabet (I, V, X, L, C, D, M) to represent numbers. For example:
1 is "I", 4 is "IV", 9 is "IX", 58 is "LVIII", and 1994 is "MCMXCIV".

This package supports conversion for integers in the range 1 to 3999.
*/
package romannumerals

import (
	"fmt"
	"strings"
)

// roman represents a Roman numeral and its corresponding integer value.
type roman struct {
	Symbol string
	Value  int
}

// Predefined Roman numeral symbols and their values, ordered from highest to lowest.
var romans = []roman{
	{"M", 1000},
	{"CM", 900},
	{"D", 500},
	{"CD", 400},
	{"C", 100},
	{"XC", 90},
	{"L", 50},
	{"XL", 40},
	{"X", 10},
	{"IX", 9},
	{"V", 5},
	{"IV", 4},
	{"I", 1},
}

// ToRomanNumeral converts an integer to its Roman numeral representation.
// Returns an error if the input is not in the range 1 to 3999.
func ToRomanNumeral(input int) (string, error) {
	var builder strings.Builder

	if input <= 0 || input >= 4000 {
		return "", fmt.Errorf("invalid value [%d]", input)
	}

	for _, roman := range romans {
		for input >= roman.Value {
			input -= roman.Value
			builder.WriteString(roman.Symbol)
		}
	}

	return builder.String(), nil
}
