package collatzconjecture

import (
	"fmt"
)

type CollazError struct {
	Value   int
	Message string
}

func (e *CollazError) Error() string {
	return fmt.Sprintf("Error with value %d: %s", e.Value, e.Message)
}

func CollatzConjecture(n int) (int, error) {
	if n < 1 {
		return 0, &CollazError{Value: n, Message: "Input must be greater than zero"}
	}

	trials := 0
	for ; n != 1; trials++ {
		if n%2 == 0 {
			n = n / 2
		} else {
			n = 3*n + 1
		}
	}

	return trials, nil
}
