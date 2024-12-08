/*
Package prime provides utilities to compute the prime factorization of integers.
*/
package prime

// Factors computes the prime factors of the given integer n.
// It returns a slice of prime factors in ascending order.
// The function iteratively divides n by its smallest factors until n is fully factorized.
func Factors(n int64) []int64 {
	factors := make([]int64, 0)

	// Even factors
	for n%2 == 0 {
		factors = append(factors, 2)
		n = n / 2
	}

	// Odd factors
	var i int64
	for i = 3; i*i <= n; i += 2 {
		for n%i == 0 {
			factors = append(factors, i)
			n = n / i
		}
	}

	// Prime factor
	if n > 1 {
		factors = append(factors, n)
	}

	return factors
}
