// Package twofer implements a utility function which determines how to answer
// when giving away an extra cookie.
package twofer

import "fmt"

// ShareWith is a function which returns a string representing waht to say
// when giving away an extra cookie. The result changes depending on whether
// we know the other person's name or not.
func ShareWith(name string) string {
	return fmt.Sprintf("One for %s, one for me.", othersName(name))
}

func othersName(name string) string {
	if name == "" {
		return "you"
	}
	return name
}
