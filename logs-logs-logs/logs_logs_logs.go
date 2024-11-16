// Package logs provides utilities for analyzing and manipulating log entries
// emitted by various applications.
package logs

import (
	"strings"
	"unicode/utf8"
)

const (
	// Application names corresponding to specific log identifiers.
	recommendationLog = "recommendation"
	searchLog         = "search"
	weatherLog        = "weather"
	defaultLog        = "default"

	// Unicode identifiers for different log types.
	recommendationLogId = '\u2757'     // ❗
	searchLogId         = '\U0001F50D' // 🔍
	weatherLogId        = '\u2600'     // ☀
)

// Application identifies the application emitting the given log entry based on specific Unicode characters.
// It returns "recommendation", "search", or "weather" if a matching identifier is found.
// If no match is found, it returns "default".
func Application(log string) string {
	for _, r := range log {
		switch r {
		case recommendationLogId:
			return recommendationLog
		case searchLogId:
			return searchLog
		case weatherLogId:
			return weatherLog
		}
	}
	return defaultLog
}

// Replace replaces all occurrences of oldRune with newRune in the provided log string.
// It returns a new string with the replacements.
func Replace(log string, oldRune, newRune rune) string {
	builder := strings.Builder{}

	for _, r := range log {
		if r == oldRune {
			builder.WriteRune(newRune)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// WithinLimit checks if the number of characters in the provided log is within the given limit.
// It returns true if the log length (in runes) is less than or equal to the limit, otherwise false.
func WithinLimit(log string, limit int) bool {
	return utf8.RuneCountInString(log) <= limit
}
