/*
Package lsproduct provides a function to calculate the largest product
of a contiguous substring of digits with a given span.
*/
package lsproduct

import (
	"fmt"
	"unicode"
)

// LargestSeriesProduct returns the largest product of a contiguous substring of digits
// with the given span. Returns an error if the span is invalid or the input contains non-digits.
func LargestSeriesProduct(digits string, span int) (int64, error) {
	if span < 0 || len(digits) < span {
		return 0, fmt.Errorf("span must be non-negative and smaller than string length")
	}

	for _, r := range digits {
		if !unicode.IsDigit(r) {
			return 0, fmt.Errorf("input contains non-digit characters")
		}
	}

	var maxProduct int64
	var currentProduct int64 = 1
	zeroCount := 0

	for i := 0; i < len(digits); i++ {
		digitValue := int64(digits[i] - '0')

		if digitValue == 0 {
			zeroCount++
		} else {
			currentProduct = currentProduct * digitValue
		}

		if i >= span {
			outgoing := int64(digits[i-span] - '0')
			if outgoing == 0 {
				zeroCount--
			} else {
				currentProduct = currentProduct / outgoing
			}
		}

		if zeroCount == 0 && i >= span-1 {
			if currentProduct > maxProduct {
				maxProduct = currentProduct
			}
		}
	}

	return maxProduct, nil
}
