// Package strain provides generic functions for filtering slices using predicates.
package strain

// predicate defines a generic function type that takes a value of type T
// and returns a boolean indicating whether the value satisfies a condition.
type predicate[T any] func(value T) bool

// Keep filters elements in the input slice based on the provided predicate function.
// It returns a new slice containing only the elements that satisfy the predicate.
func Keep[T any](values []T, p predicate[T]) []T {
	filtered := make([]T, 0)
	for _, value := range values {
		if p(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

// Discard filters elements in the input slice that do not satisfy the given predicate.
// It reuses the Keep function by negating the predicate.
func Discard[T any](values []T, p predicate[T]) []T {
	return Keep(values, func(value T) bool {
		return !p(value)
	})
}
