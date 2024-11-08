// Package blackjack provides utilities to determine actions
// in the game of Blackjack.
package blackjack

// ParseCard returns the integer value of a card following the blackjack ruleset.
// Cards "two" through "ten" have their respective values. The "ace" is worth 11,
// face cards ("jack", "queen", "king") are worth 10, and the "joker" has a value of 0.
func ParseCard(card string) int {
	switch card {
	case "ace":
		return 11
	case "two":
		return 2
	case "three":
		return 3
	case "four":
		return 4
	case "five":
		return 5
	case "six":
		return 6
	case "seven":
		return 7
	case "eight":
		return 8
	case "nine":
		return 9
	case "ten", "jack", "queen", "king":
		return 10
	default:
		return 0
	}
}

const (
	split = "P"
	win   = "W"
	hit   = "H"
	stand = "S"
)

// The possible decisions are:
// - "P" (split) if the player has two aces.
// - "W" (win) if the player's total is 21 and the dealer's card is less than 10.
// - "H" (hit) if the player's total is between 12 and 16 (inclusive) and the dealer's card is 7 or higher.
// - "S" (stand) if the player's total is between 12 and 16 (inclusive) and the dealer's card is less than 7.
// - "H" (hit) if the player's total is 11 or less.
// - "S" (stand) for any other totals.
func FirstTurn(card1, card2, dealerCard string) string {
	handValue := ParseCard(card1) + ParseCard(card2)
	dealerValue := ParseCard(dealerCard)

	switch {
	case card1 == "ace" && card2 == "ace":
		return split
	case handValue == 21 && dealerValue < 10:
		return win
	case handValue >= 12 && handValue <= 16 && dealerValue >= 7 || handValue <= 11:
		return hit
	default:
		return stand
	}
}
