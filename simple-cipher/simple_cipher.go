/*
Package cipher provides implementations of Caesar, Shift, and Vigenère ciphers.
These ciphers are used to encode and decode strings by shifting the characters
of the alphabet based on specific rules.

The Caesar cipher is a special case of the Shift cipher with a fixed shift of 3.
The Shift cipher allows for an arbitrary shift distance. The Vigenère cipher uses
a keyword to cyclically shift characters by varying amounts.
*/
package cipher

import (
	"strings"
	"unicode"
)

// shift represents a Shift cipher with a specified distance.
type shift struct {
	distance int
}

// vigenere represents a Vigenère cipher with a given key.
type vigenere struct {
	key string
}

const caesarShift = 3

// NewCaesar creates a Caesar cipher with a fixed shift of 3.
func NewCaesar() Cipher {
	return shift{
		distance: caesarShift,
	}
}

// NewShift creates a Shift cipher with the specified distance.
// The distance must be in the range 1 to 25 or -1 to -25. A distance of 0 is not allowed.
// Returns nil if the distance is outside the valid range.
func NewShift(distance int) Cipher {
	if (distance < -25 || distance > 25) || distance == 0 {
		return nil
	}
	return shift{
		distance: distance,
	}
}

// Encode encodes the input string using the Shift cipher.
func (c shift) Encode(input string) string {
	return encode(input, c.distance)
}

// Decode decodes the input string using the Shift cipher.
func (c shift) Decode(input string) string {
	return encode(input, -c.distance)
}

// encode shifts the characters in the input string by the specified distance.
func encode(input string, shift int) string {
	builder := strings.Builder{}

	for _, r := range strings.ToLower(input) {
		if unicode.IsLower(r) {
			encoded := 'a' + (r-'a'+rune(shift))%26
			if encoded < 'a' {
				encoded += 26
			}
			builder.WriteRune(encoded)
		}
	}

	return builder.String()
}

// NewVigenere creates a Vigenère cipher with the specified key.
// Returns nil if the key is invalid (empty, non-lowercase, or all 'a's).
func NewVigenere(key string) Cipher {
	if len(key) == 0 || len(strings.Trim(key, "a")) == 0 {
		return nil
	}
	for _, r := range key {
		if !unicode.IsLower(r) {
			return nil
		}
	}
	return vigenere{
		key: key,
	}
}

// Encode encodes the input string using the Vigenère cipher.
func (v vigenere) Encode(input string) string {
	builder := strings.Builder{}
	keyLen := len(v.key)
	keyIndex := 0

	for _, r := range strings.ToLower(input) {
		if unicode.IsLower(r) {
			shift := v.key[keyIndex%keyLen] - 'a'
			encoded := 'a' + (r-'a'+rune(shift))%26
			builder.WriteRune(encoded)
			keyIndex++
		}
	}

	return builder.String()
}

// Decode decodes the input string using the Vigenère cipher.
func (v vigenere) Decode(input string) string {
	builder := strings.Builder{}
	keyLen := len(v.key)
	keyIndex := 0

	for _, r := range strings.ToLower(input) {
		if unicode.IsLower(r) {
			shift := v.key[keyIndex%keyLen] - 'a'
			decoded := 'a' + (r-'a'-rune(shift)+26)%26
			builder.WriteRune(decoded)
			keyIndex++
		}
	}

	return builder.String()
}
