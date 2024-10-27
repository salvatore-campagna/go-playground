// Package cars provides functions to calculate various production metrics
// for a car manufacturing assembly line.
//
// The package includes:
//
//   - CalculateWorkingCarsPerHour: Determines the number of functional cars produced per hour,
//     based on production rate and success rate.
//   - CalculateWorkingCarsPerMinute: Computes the number of functional cars produced per minute,
//     based on the hourly rate and success rate.
//   - CalculateCost: Calculates the production cost for a given number of cars, where groups
//     of 10 cars are cheaper to produce than single units.
//
// This package is intended to help analyze and estimate car production efficiency
// and costs on an assembly line.
package cars

// CalculateWorkingCarsPerHour calculates how many working cars are
// produced by the assembly line every hour.
func CalculateWorkingCarsPerHour(productionRate int, successRate float64) float64 {
	return float64(productionRate) * successRate / 100.0
}

// CalculateWorkingCarsPerMinute calculates how many working cars are
// produced by the assembly line every minute.
func CalculateWorkingCarsPerMinute(productionRate int, successRate float64) int {
	return int(float64(productionRate) * successRate / 100.0 / 60.0)
}

// CalculateCost works out the cost of producing the given number of cars.
func CalculateCost(carsCount int) uint {
	tens := carsCount / 10
	ones := carsCount % 10

	return uint(tens*95_000 + ones*10_000)
}
