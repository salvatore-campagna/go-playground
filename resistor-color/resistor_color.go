// Package resistorcolor provides functionality to work with resistor color codes.
// Each color represents a numeric value used to determine the resistance of a resistor.
// The package allows querying the full list of colors and retrieving the numeric value for a specific color.
package resistorcolor

// colors defines the standard list of resistor colors in ascending order of their numeric values.
var colors = []string{
	"black",
	"brown",
	"red",
	"orange",
	"yellow",
	"green",
	"blue",
	"violet",
	"grey",
	"white",
}

// invalidColor is the return value for an invalid or unrecognized color.
// It indicates that the provided color does not correspond to a valid resistor color.
const invalidColor = -1

// colorCodes is a precomputed map for efficient lookups of numeric values for resistor colors.
// The keys are color names, and the values are the corresponding numeric values.
var colorCodes = func() map[string]int {
	m := make(map[string]int)
	for i, c := range colors {
		m[c] = i
	}
	return m
}()

// Colors returns the list of all resistor colors in order of their numeric values.
// This can be used to retrieve the complete set of standard resistor colors.
func Colors() []string {
	return colors
}

// ColorCode returns the numeric value of the given resistor color.
// If the provided color is not valid, it returns `invalidColor`.
//
// Example:
//
//	value := ColorCode("red")  // returns 2
//	value := ColorCode("gold") // returns -1 (invalid color)
func ColorCode(color string) int {
	if colorCode, exists := colorCodes[color]; exists {
		return colorCode
	}
	return invalidColor
}
