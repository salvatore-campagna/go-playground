/*
Package cards provides functions for manipulating slices of integers,
representing a collection of cards. It includes functions to retrieve, modify,
prepend, and remove items, as well as to retrieve specific favorite cards.
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
	return append(copy, slice...)
}

// RemoveItem removes an item from a slice by modifying the existing slice.
func RemoveItem(slice []int, index int) []int {
	if index < 0 || index >= len(slice) {
		return append(slice, []int{}...)
	}

	b := make([]int, 0, len(slice)-1)
	b = append(b, slice[:index]...)
	return append(b, slice[index+1:]...)
}
