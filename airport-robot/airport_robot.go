/*
Package airportrobot provides functionality for simulating a multilingual greeting system at an airport.
It uses the Greeter interface to support multiple languages, allowing greetings to be customized per language.
*/
package airportrobot

import "fmt"

// Greeter is an interface for creating multilingual greetings.
// It defines methods to get the name of the language and generate a greeting for a visitor.
type Greeter interface {
	LanguageName() string
	Greet(visitorName string) string
}

// SayHello generates a multilingual greeting for a visitor.
// It includes the language name and the greeting text provided by the Greeter.
func SayHello(visitorName string, g Greeter) string {
	return fmt.Sprintf("I can speak %s: %s", g.LanguageName(), g.Greet(visitorName))
}

// Italian is a struct representing a greeter that speaks Italian.
type Italian struct{}

// LanguageName returns the name of the Italian language.
func (i Italian) LanguageName() string {
	return "Italian"
}

// Greet generates a greeting in Italian for the given visitor.
func (i Italian) Greet(visitorName string) string {
	return fmt.Sprintf("Ciao %s!", visitorName)
}

// Portuguese is a struct representing a greeter that speaks Portuguese.
type Portuguese struct{}

// LanguageName returns the name of the Portuguese language.
func (p Portuguese) LanguageName() string {
	return "Portuguese"
}

// Greet generates a greeting in Portuguese for the given visitor.
func (p Portuguese) Greet(visitorName string) string {
	return fmt.Sprintf("Olá %s!", visitorName)
}
