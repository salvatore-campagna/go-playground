// Package interest provides functions for calculating interest rates, annual balance updates,
// and estimating the number of years needed to reach a desired balance.
package interest

const (
	negativeBalanceRate = 3.213
	lowBalanceRate      = 0.5
	mediumBalanceRate   = 1.621
	highBalanceRate     = 2.475

	lowBalanceThreshold    = 0.0
	mediumBalanceThreshold = 1000.0
	highBalanceThreshold   = 5000.0
)

// InterestRate returns the interest rate for the provided balance.
// The interest rate is determined based on the following conditions:
// - Negative balance: 3.213%
// - Balance between 0 and 999.99: 0.5%
// - Balance between 1000 and 4999.99: 1.621%
// - Balance of 5000 or more: 2.475%
func InterestRate(balance float64) float32 {
	switch {
	case balance < 0:
		return negativeBalanceRate
	case balance < mediumBalanceThreshold:
		return lowBalanceRate
	case balance < highBalanceThreshold:
		return mediumBalanceRate
	default:
		return highBalanceRate
	}
}

// Interest calculates the interest for the provided balance based on its interest rate.
// The formula used is: balance * (interest rate / 100).
func Interest(balance float64) float64 {
	return balance * float64(InterestRate(balance)) * 0.01
}

// AnnualBalanceUpdate calculates the annual balance update, taking into account the interest rate.
// It returns the new balance after adding the calculated interest to the original balance.
func AnnualBalanceUpdate(balance float64) float64 {
	return balance + Interest(balance)
}

// YearsBeforeDesiredBalance calculates the minimum number of years required to reach the desired balance.
// It takes the initial balance and the target balance as inputs and returns the number of years required
// for the balance to meet or exceed the target, based on the annual balance update.
func YearsBeforeDesiredBalance(balance, targetBalance float64) int {
	yearsToTarget := 0
	for ; balance < targetBalance; yearsToTarget++ {
		balance = AnnualBalanceUpdate(balance)
	}
	return yearsToTarget
}
