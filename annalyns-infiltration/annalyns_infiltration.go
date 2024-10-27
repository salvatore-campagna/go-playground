// Package annalyn provides functions for assessing the conditions under which
// Annalyn can perform certain actions in her infiltration mission.
//
// The package includes the following capabilities:
//
// - Determining if Annalyn can perform a fast attack on the knight, based on the knight's status.
// - Checking if Annalyn can spy on characters by evaluating their states of wakefulness.
// - Assessing whether Annalyn can signal to the prisoner given the archer's and prisoner's states.
// - Determining if Annalyn can free the prisoner based on various conditions, including the presence of her pet dog.
//
// Each function represents a specific condition that must be met to execute an action,
// supporting Annalyn's mission strategy.
package annalyn

// CanFastAttack can be executed only when the knight is sleeping.
func CanFastAttack(knightIsAwake bool) bool {
	return !knightIsAwake
}

// CanSpy can be executed if at least one of the characters is awake.
func CanSpy(knightIsAwake, archerIsAwake, prisonerIsAwake bool) bool {
	return knightIsAwake || archerIsAwake || prisonerIsAwake
}

// CanSignalPrisoner can be executed if the prisoner is awake and the archer is sleeping.
func CanSignalPrisoner(archerIsAwake, prisonerIsAwake bool) bool {
	return !archerIsAwake && prisonerIsAwake
}

// CanFreePrisoner can be executed if the prisoner is awake and the other 2 characters are asleep
// or if Annalyn's pet dog is with her and the archer is sleeping.
func CanFreePrisoner(knightIsAwake, archerIsAwake, prisonerIsAwake, petDogIsPresent bool) bool {
	return (prisonerIsAwake && !knightIsAwake && !archerIsAwake) || (petDogIsPresent && !archerIsAwake)
}
