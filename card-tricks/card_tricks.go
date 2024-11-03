/*
Package cards provides functions for manipulating slices of integers,
representing a collection of cards. It includes functions to retrieve, modify,
prepend, and remove items, as well as to retrieve specific favorite cards.

Types:
  - FavoriteCards: Returns a predefined slice of favorite cards in a specified order.
  - GetItem: Retrieves an item from a slice at a given position, returning -1 if out of range.
  - SetItem: Sets or appends an item in a slice at a specified position.
  - PrependItems: Prepends a variable number of items to the beginning of a slice.
  - RemoveItem: Removes an item from a slice at a specified position, returning the original slice if out of range.
*/
package cards

// FavoriteCards returns a slice with the cards 2, 6 and 9 in that order.
func FavoriteCards() []int {
	return []int{2, 6, 9}
}

// GetItem retrieves an item from a slice at given position.
// If the index is out of range, we want it to return -1.
func GetItem(slice []int, index int) int {
	if index < 0 || index >= len(slice) {
		return -1
	}

	return slice[index]
}

// SetItem writes an item to a slice at given position overwriting an existing value.
// If the index is out of range the value needs to be appended.
func SetItem(slice []int, index, value int) []int {
	if index < 0 || index >= len(slice) {
		return append(slice, value)
	}

	slice[index] = value
	return slice
}

// PrependItems adds an arbitrary number of values at the front of a slice.
func PrependItems(slice []int, values ...int) []int {
	copy := make([]int, 0, len(values)+len(slice))
	copy = append(copy, values...)
	copy = append(copy, slice...)
	return copy
}

// RemoveItem removes an item from a slice by modifying the existing slice.
func RemoveItem(slice []int, index int) []int {
	if index < 0 || index >= len(slice) {
		return append(slice, []int{}...)
	}

	b := make([]int, 0, len(slice)-1)
	b = append(b, slice[:index]...)
	b = append(b, slice[index+1:]...)
	return b
}
