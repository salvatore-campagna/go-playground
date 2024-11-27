package resistorcolorduo

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

// colorCodes is a precomputed map for efficient lookups of numeric values for resistor colors.
// The keys are color names, and the values are the corresponding numeric values.
var colorCodes = func() map[string]int {
	m := make(map[string]int)
	for i, c := range colors {
		m[c] = i
	}
	return m
}()

// Value returns the resistance value of a resistor based on the first two color bands.
func Value(colors []string) int {
	return colorCodes[colors[0]]*10 + colorCodes[colors[1]]
}
