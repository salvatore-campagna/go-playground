/*
Package anagram provides utilities to detect anagrams from a list of words.
An anagram is a word formed by rearranging the letters of another word.
*/
package anagram

import (
	"sort"
	"strings"
)

/*
Detect identifies all anagrams of the given subject from a list of candidates.
It compares the sorted characters of the subject with each candidate, ignoring case.
Returns a slice of strings containing the detected anagrams.
*/
func Detect(subject string, candidates []string) []string {
	sortedSubject := sortString(strings.ToLower(subject))
	anagrams := make([]string, 0)

	for _, candidate := range candidates {
		if strings.EqualFold(candidate, subject) {
			continue
		}

		sortedCandidate := sortString(strings.ToLower(candidate))
		if sortedCandidate == sortedSubject {
			anagrams = append(anagrams, candidate)
		}
	}

	return anagrams
}

/*
sortString sorts the characters of a string alphabetically.
It returns the sorted string and is used to compare words for anagram detection.
*/
func sortString(input string) string {
	runes := []rune(input)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}
