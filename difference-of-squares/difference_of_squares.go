// Package diffsquares provides functions to compute the square of the sum,
// the sum of the squares, and the difference between the two for a given range of numbers.
package diffsquares

// SquareOfSum calculates the square of the sum of the first n natural numbers.
// It returns the square of the sum: (1 + 2 + ... + n)^2.
func SquareOfSum(n int) int {
	sum := n * (n + 1) / 2
	return sum * sum
}

// SumOfSquares calculates the sum of the squares of the first n natural numbers.
// It returns the sum of squares: 1^2 + 2^2 + ... + n^2.
func SumOfSquares(n int) int {
	return n * (n + 1) * (2*n + 1) / 6
}

// Difference calculates the difference between the square of the sum
// and the sum of the squares of the first n natural numbers.
// It returns SquareOfSum(n) - SumOfSquares(n).
func Difference(n int) int {
	return SquareOfSum(n) - SumOfSquares(n)
}
