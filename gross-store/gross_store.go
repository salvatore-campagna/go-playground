// Package gross provides functions to manage items and their quantities
// in a grocery store billing system.
package gross

// Constants representing the quantities for different units of measurement.
const (
	quarterOfADozenQuantity = 3
	halfOfADozenQuantity    = 6
	dozenQuantity           = 12
	smallGrossQuantity      = 120
	grossQuantity           = 144
	greatGrossQuantity      = 1728
)

// Units returns a map of predefined unit measurements used in the Gross Store.
// The map contains units such as "dozen", "gross", and "great gross" with their respective quantities.
func Units() map[string]int {
	return map[string]int{
		"quarter_of_a_dozen": quarterOfADozenQuantity,
		"half_of_a_dozen":    halfOfADozenQuantity,
		"dozen":              dozenQuantity,
		"small_gross":        smallGrossQuantity,
		"gross":              grossQuantity,
		"great_gross":        greatGrossQuantity,
	}
}

// NewBill creates and returns a new, empty bill for storing item quantities.
func NewBill() map[string]int {
	return map[string]int{}
}

// AddItem adds an item to the customer's bill with the specified unit quantity.
// If the unit is not recognized, it returns false. Otherwise, it adds the quantity to the bill.
func AddItem(bill, units map[string]int, item, unit string) bool {
	unitQuantity, exists := units[unit]
	if !exists {
		return false
	}
	bill[item] += unitQuantity
	return true
}

// RemoveItem removes a specified quantity of an item from the customer's bill.
// If the item or unit doesn't exist, or if the quantity to remove is greater than what is in the bill,
// it returns false. If the remaining quantity is zero, the item is removed from the bill.
func RemoveItem(bill, units map[string]int, item, unit string) bool {
	billQuantity, itemExists := bill[item]
	if !itemExists {
		return false
	}

	unitQuantity, unitExists := units[unit]
	if !unitExists {
		return false
	}

	if billQuantity < unitQuantity {
		return false
	}

	remainingQuantity := billQuantity - unitQuantity
	if remainingQuantity == 0 {
		delete(bill, item)
		return true
	}

	bill[item] = remainingQuantity
	return true
}

// GetItem returns the quantity of a specified item from the customer's bill.
// If the item is not found, it returns 0 and false.
func GetItem(bill map[string]int, item string) (int, bool) {
	quantity, itemExists := bill[item]
	return quantity, itemExists
}
