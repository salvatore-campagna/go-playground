// Package purchase provides functions to determine whether a license is
// needed for specific vehicle types, recommend between two vehicles,
// and calculate the resell price based on vehicle age.
package purchase

// NeedsLicense determines whether a license is needed to drive a type of vehicle. Only "car" and "truck" require a license.
func NeedsLicense(kind string) bool {
	switch kind {
	case "car", "truck":
		return true
	default:
		return false
	}
}

// ChooseVehicle recommends a vehicle for selection. It always recommends the vehicle that comes first in lexicographical order.
func ChooseVehicle(option1, option2 string) string {
	switch {
	case option1 <= option2:
		return option1 + " is clearly the better choice."
	default:
		return option2 + " is clearly the better choice."
	}
}

// CalculateResellPrice calculates how much a vehicle can resell for at a certain age.
func CalculateResellPrice(originalPrice, age float64) float64 {
	switch {
	case age >= 10:
		return originalPrice * 0.5
	case age < 3:
		return originalPrice * 0.8
	default:
		return originalPrice * 0.7
	}
}
