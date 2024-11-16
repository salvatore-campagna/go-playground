// Package lasagna provides functions to assist with preparing lasagna recipes.
package lasagna

import "strings"

// Constants for preparation and ingredient quantities per layer.
const (
	defaultAverageTimePerLayer = 2   // Default preparation time per layer in minutes
	noodlesPerLayer            = 50  // Grams of noodles per layer
	saucePerLayer              = 0.2 // Liters of sauce per layer
)

// PreparationTime calculates the total preparation time for the given layers.
// If averageTimePerLayer is 0, it uses the default time of 2 minutes per layer.
func PreparationTime(layers []string, averageTimePerLayer int) int {
	if averageTimePerLayer == 0 {
		averageTimePerLayer = defaultAverageTimePerLayer
	}
	return len(layers) * averageTimePerLayer
}

// Quantities calculates the total quantities of noodles (in grams) and sauce (in liters)
// required for the given layers.
func Quantities(layers []string) (int, float64) {
	totalNoodlesAmount := 0
	totalSauceAmount := 0.0

	for _, layer := range layers {
		switch strings.ToLower(layer) {
		case "noodles":
			totalNoodlesAmount += noodlesPerLayer
		case "sauce":
			totalSauceAmount += saucePerLayer
		}
	}

	return totalNoodlesAmount, totalSauceAmount
}

// AddSecretIngredient takes your friend's ingredients and adds their secret ingredient
// to your list of ingredients by replacing the last element of your ingredients.
func AddSecretIngredient(friendsIngredients []string, myIngredients []string) {
	myIngredients[len(myIngredients)-1] = friendsIngredients[len(friendsIngredients)-1]
}

// ScaleRecipe scales the quantities for the desired number of portions.
// The quantities are scaled relative to a base of 2 portions.
func ScaleRecipe(quantities []float64, numberOfPortions int) []float64 {
	scaledQuantities := make([]float64, len(quantities))
	scaleFactor := float64(numberOfPortions) / 2.0
	for i, quantity := range quantities {
		scaledQuantities[i] = quantity * scaleFactor
	}
	return scaledQuantities
}
