/*
Package triangle provides utilities to determine the type of a triangle
(equilateral, isosceles, or scalene) based on the lengths of its sides.
*/
package triangle

// Kind represents the type of a triangle.
type Kind int

const (
	NaT = iota // Not a triangle
	Equ        // Equilateral
	Iso        // Isosceles
	Sca        // Scalene
)

/*
KindFromSides determines the type of triangle given the lengths of its three sides.

Returns:
  - NaT: Not a triangle if the sides do not satisfy triangle properties.
  - Equ: Equilateral triangle if all sides are equal.
  - Iso: Isosceles triangle if two sides are equal.
  - Sca: Scalene triangle if all sides are different.
*/
func KindFromSides(a, b, c float64) Kind {
	if !isTriangle(a, b, c) {
		return NaT
	}

	if isEqu(a, b, c) {
		return Equ
	} else if isIso(a, b, c) {
		return Iso
	}

	return Sca
}

// isTriangle checks if the given sides form a valid triangle.
func isTriangle(a, b, c float64) bool {
	return a > 0 && b > 0 && c > 0 && a+b >= c && a+c >= b && b+c >= a
}

// isEqu checks if the triangle is equilateral.
func isEqu(a, b, c float64) bool {
	return a == b && b == c
}

// isIso checks if the triangle is isosceles.
func isIso(a, b, c float64) bool {
	return a == b || a == c || b == c
}
