// Package darts provides a function to calculate the score of a dart
// based on its coordinates on a dartboard.
package darts

import "math"

const (
	innerCircleRadius  = 1.0
	middleCircleRadius = 5.0
	outerCircleRadius  = 10.0

	innerCirclePoints  = 10
	middleCirclePoints = 5
	outerCirclePoints  = 1
	missPoints         = 0
)

// Score calculates the score for a dart based on its (x, y) coordinates.
// The dartboard is divided into regions with different scores:
// - 10 points for a distance of 1 or less from the center
// - 5 points for a distance greater than 1 but less than or equal to 5
// - 1 point for a distance greater than 5 but less than or equal to 10
// - 0 points for a distance greater than 10
func Score(x, y float64) int {
	distance := distanceFromCenter(x, y)
	switch {
	case distance <= innerCircleRadius:
		return innerCirclePoints
	case distance <= middleCircleRadius:
		return middleCirclePoints
	case distance <= outerCircleRadius:
		return outerCirclePoints
	default:
		return missPoints
	}
}

// distanceFromCenter calculates the Euclidean distance from the center (0, 0)
// to the point (x, y) using the Pythagorean theorem.
func distanceFromCenter(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}
