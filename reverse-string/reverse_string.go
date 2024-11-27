// Package reverse provides functionality to reverse strings while properly handling Unicode characters.
package reverse

import (
	"strings"
)

// Reverse returns the input string reversed.
// It converts the string to a slice of runes to handle multi-byte characters correctly.
func Reverse(input string) string {
	runes := []rune(input)
	sb := strings.Builder{}
	for i := len(runes) - 1; i >= 0; i-- {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}
