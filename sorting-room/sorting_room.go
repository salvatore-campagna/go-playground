/*
Package sorting provides utilities for describing and extracting numbers from various types
of number-related interfaces, including NumberBox and FancyNumberBox.
*/
package sorting

import (
	"fmt"
	"strconv"
)

// DescribeNumber returns a string describing the provided float64 number.
func DescribeNumber(f float64) string {
	return fmt.Sprintf("This is the number %.1f", f)
}

// NumberBox is an interface representing a box containing an integer number.
type NumberBox interface {
	Number() int
}

// DescribeNumberBox returns a string describing the NumberBox.
func DescribeNumberBox(nb NumberBox) string {
	return fmt.Sprintf("This is a box containing the number %.1f", float64(nb.Number()))
}

// FancyNumber is a struct representing a fancy number stored as a string.
type FancyNumber struct {
	n string
}

// Value returns the string representation of the FancyNumber.
func (i FancyNumber) Value() string {
	return i.n
}

// FancyNumberBox is an interface representing a box containing a fancy number.
type FancyNumberBox interface {
	Value() string
}

// ExtractFancyNumber returns the integer value of a FancyNumberBox if it is of type FancyNumber,
// or 0 if it is of any other type.
func ExtractFancyNumber(fnb FancyNumberBox) int {
	switch fnb.(type) {
	case FancyNumber:
		value, err := strconv.Atoi(fnb.Value())
		if err != nil {
			return 0
		}
		return value
	default:
		return 0
	}
}

// DescribeFancyNumberBox returns a string describing the FancyNumberBox.
func DescribeFancyNumberBox(fnb FancyNumberBox) string {
	return fmt.Sprintf("This is a fancy box containing the number %.1f", float64(ExtractFancyNumber(fnb)))
}

// DescribeAnything returns a string describing the provided interface.
// It handles integers, floats, NumberBox, FancyNumberBox, and defaults to a generic message.
func DescribeAnything(i interface{}) string {
	switch i := i.(type) {
	case int:
		return DescribeNumber(float64(i))
	case float64:
		return DescribeNumber(i)
	case NumberBox:
		return DescribeNumberBox(i)
	case FancyNumberBox:
		return DescribeFancyNumberBox(i)
	default:
		return "Return to sender"
	}
}
