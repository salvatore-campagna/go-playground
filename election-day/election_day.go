/*
Package electionday provides utilities for managing election votes and results.
It includes functionality for creating vote counters, updating vote counts,
and handling election results.

Functions:
  - NewVoteCounter: Creates a new vote counter with initial votes.
  - VoteCount: Extracts the number of votes from a counter.
  - IncrementVoteCount: Increments the vote count in a counter.
  - NewElectionResult: Creates a new election result with a candidate's name and vote count.
  - DisplayResult: Formats and returns a string representation of an election result.
  - DecrementVotesOfCandidate: Decrements the vote count of a specific candidate in a map.

Types:
  - ElectionResult: Represents the result of an election for a specific candidate.
*/

package electionday

import "fmt"

// NewVoteCounter returns a new vote counter with
// a given number of initial votes.
func NewVoteCounter(initialVotes int) *int {
	return &initialVotes
}

// VoteCount extracts the number of votes from a counter.
// If the counter is nil, it returns 0.
func VoteCount(counter *int) int {
	if counter == nil {
		return 0
	}
	return *counter
}

// IncrementVoteCount increments the value in a vote counter
// by the specified increment.
func IncrementVoteCount(counter *int, increment int) {
	*counter += increment
}

// NewElectionResult creates a new election result with the
// candidate's name and vote count.
func NewElectionResult(candidateName string, votes int) *ElectionResult {
	return &ElectionResult{
		Name:  candidateName,
		Votes: votes,
	}
}

// DisplayResult returns a formatted string representing
// the election result in the format: "<Name> (<Votes>)".
func DisplayResult(result *ElectionResult) string {
	return fmt.Sprintf("%s (%d)", result.Name, result.Votes)
}

// DecrementVotesOfCandidate decrements the vote count
// of the specified candidate in the results map by 1.
func DecrementVotesOfCandidate(results map[string]int, candidate string) {
	results[candidate] = results[candidate] - 1
}
