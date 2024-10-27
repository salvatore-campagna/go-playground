package fizzbuzz

import (
	"strconv"
	"strings"
)

func FizzBuzz(n int) string {
	var sb strings.Builder

	for i := 1; i <= n; i++ {
		length := sb.Len()
		if i%3 == 0 {
			sb.WriteString("Fizz")
		}
		if i%5 == 0 {
			sb.WriteString("Buzz")
		}
		if sb.Len() == length {
			sb.WriteString(strconv.Itoa(i))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
