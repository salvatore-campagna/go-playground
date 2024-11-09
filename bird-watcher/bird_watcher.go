// Package birdwatcher provides functions to analyze bird counts recorded over a number of days.
// It includes utilities to calculate total counts, weekly counts, and to fix logs with missing data.
package birdwatcher

// TotalBirdCount returns the total bird count by summing the individual day's counts.
// It takes a slice of integers, where each element represents the count of birds observed on a given day.
func TotalBirdCount(birdsPerDay []int) int {
	totalBirds := 0
	for _, count := range birdsPerDay {
		totalBirds += count
	}
	return totalBirds
}

const daysPerWeek = 7

// BirdsInWeek returns the total bird count for a given week by summing the counts of that specific week.
// The parameter `week` is 1-indexed (i.e., week 1 refers to days 1-7, week 2 refers to days 8-14, etc.).
func BirdsInWeek(birdsPerDay []int, week int) int {
	if week <= 0 || len(birdsPerDay) < daysPerWeek*week {
		// Normally we would return an error but the function expects retruning only an `int`
		return 0
	}
	start := daysPerWeek * (week - 1)
	end := start + daysPerWeek
	return TotalBirdCount(birdsPerDay[start:end])
}

// FixBirdCountLog returns the bird counts after correcting the logs for alternate days.
// It increases the count by 1 for every other day, starting from day 0 (i.e., the first day).
// A copy of the original slice is returned instead of modifying the input slice in place.
func FixBirdCountLog(birdsPerDay []int) []int {
	fixedBirdsPerDay := make([]int, len(birdsPerDay))
	copy(fixedBirdsPerDay, birdsPerDay)
	for i := 0; i < len(fixedBirdsPerDay); i += 2 {
		fixedBirdsPerDay[i]++
	}
	return fixedBirdsPerDay
}
