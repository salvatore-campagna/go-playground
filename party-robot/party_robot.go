// Package partyrobot provides functions to create personalized greeting messages
// for a party. It can generate a welcome message, a birthday wish, and assign a
// specific table with detailed information for each guest.
package partyrobot

import "fmt"

// Welcome greets a person by name.
func Welcome(name string) string {
	return fmt.Sprintf("Welcome to my party, %s!", name)
}

// HappyBirthday wishes happy birthday to the birthday person and exclaims their age.
func HappyBirthday(name string, age int) string {
	return fmt.Sprintf("Happy birthday %s! You are now %v years old!", name, age)
}

// AssignTable assigns a table to each guest with a formatted welcome message.
func AssignTable(name string, table int, neighbor, direction string, distance float64) string {
	return fmt.Sprintf("%s\nYou have been assigned to table %03d. Your table is %s, exactly %.1f meters from here.\nYou will be sitting next to %s.",
		Welcome(name), table, direction, distance, neighbor)
}
