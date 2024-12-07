/*
Package sieve provides an implementation of the Sieve of Eratosthenes
algorithm to compute prime numbers up to a specified limit.
*/
package sieve

/*
Sieve computes all prime numbers up to a given limit using the
Sieve of Eratosthenes algorithm.
*/
func Sieve(limit int) []int {
	if limit < 2 {
		return []int{}
	}

	values := make([]bool, limit+1)
	for i := 2; i*i <= limit; i++ {
		if !values[i] {
			for j := i * i; j <= limit; j += i {
				values[j] = true
			}
		}
	}

	result := make([]int, 0, limit)
	for i := 2; i <= limit; i++ {
		if !values[i] {
			result = append(result, i)
		}
	}

	return result
}
