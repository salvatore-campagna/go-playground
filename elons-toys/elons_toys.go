/*
Package elon provides a simulation of a toy car's battery usage and distance tracking.
It includes methods for driving the car, displaying its status, and determining if it can complete a race.
*/
package elon

import "fmt"

// Drive simulates the car driving one step.
// It reduces the battery by the battery drain and increases the distance by the speed,
// provided there is enough battery remaining.
func (c *Car) Drive() {
	if c.battery > c.batteryDrain {
		c.battery -= c.batteryDrain
		c.distance += c.speed
	}
}

// DisplayDistance returns the distance driven as a formatted string.
func (c *Car) DisplayDistance() string {
	return fmt.Sprintf("Driven %d meters", c.distance)
}

// DisplayBattery returns the remaining battery percentage as a formatted string.
func (c *Car) DisplayBattery() string {
	return fmt.Sprintf("Battery at %d%%", c.battery)
}

// CanFinish checks if the car can complete a given track distance with the remaining battery.
func (c *Car) CanFinish(trackDistance int) bool {
	maxDistance := (c.battery / c.batteryDrain) * c.speed
	return maxDistance >= trackDistance
}
