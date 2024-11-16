// Package leap provides a function to determine if a given year is a leap year.
package leap

// IsLeapYear determines if a given year is a leap year based on the following rules:
// - A year is a leap year if it is divisible by 400.
// - A year is also a leap year if it is divisible by 4 but not by 100.
// - All other years are not leap years.
func IsLeapYear(year int) bool {
	return (year%400 == 0) || (year%4 == 0 && year%100 != 0)
}
