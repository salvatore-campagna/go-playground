// Package fizzbuzz provides a function to generate a sequence of FizzBuzz values
// up to a specified number. The classic FizzBuzz problem replaces multiples of 3
// with "Fizz", multiples of 5 with "Buzz", and multiples of both with "FizzBuzz".
package fizzbuzz

import (
	"strconv"
	"strings"
)

// FizzBuzz generates a FizzBuzz sequence up to the given number n.
// For each integer from 1 to n:
// - If the number is divisible by 3, "Fizz" is added to the output.
// - If the number is divisible by 5, "Buzz" is added.
// - If the number is divisible by both 3 and 5, "FizzBuzz" is added.
// - Otherwise, the number itself is added as a string.
// Each entry in the sequence is followed by a newline character.
func FizzBuzz(n int) string {
	var sb strings.Builder

	for i := 1; i <= n; i++ {
		switch {
		case i%3 == 0 && i%5 == 0:
			sb.WriteString("FizzBuzz")
		case i%3 == 0:
			sb.WriteString("Fizz")
		case i%5 == 0:
			sb.WriteString("Buzz")
		default:
			sb.WriteString(strconv.Itoa(i))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
