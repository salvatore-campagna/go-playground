// Package rotationalcipher provides functionality to encrypt strings using a rotational (Caesar) cipher.
package rotationalcipher

import (
	"strings"
	"unicode"
)

// RotationalCipher encrypts the input string by shifting its letters by the given shift key.
// Non-alphabetic characters are not affected.
//
// Example:
//
//	RotationalCipher("Hello, World!", 5) // Output: "Mjqqt, Btwqi!"
func RotationalCipher(plain string, shiftKey int) string {
	sb := strings.Builder{}
	shiftKey = shiftKey % 26

	for _, char := range plain {
		if unicode.IsUpper(char) {
			sb.WriteRune('A' + (char-'A'+rune(shiftKey))%26)
		} else if unicode.IsLower(char) {
			sb.WriteRune('a' + (char-'a'+rune(shiftKey))%26)
		} else {
			sb.WriteRune(char)
		}
	}

	return sb.String()
}
