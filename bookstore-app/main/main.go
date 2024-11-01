package main

import (
	"bookstore-app/bookstore"
	"fmt"
)

func main() {
	// Initialize a new BookStore
	store := bookstore.NewBookStore("The Go Bookstore")

	// Add books to the store
	store.AddBook(bookstore.NewBook("The Go Programming Language", "Alan A. A. Donovan", 39.99, 5))
	store.AddBook(bookstore.NewBook("Learning Go", "Jon Bodner", 29.99, 3))
	store.AddBook(bookstore.NewBook("Introducing Go", "Caleb Doxsey", 24.99, 2))

	// Display the total inventory value
	fmt.Printf("Total inventory value: $%.2f\n", store.TotalInventoryValue())

	// Search for a book by title
	title := "The Go Programming Language"
	book, err := store.FindBookByTitle(title)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Found book: %s by %s - Price: $%.2f, In Stock: %d\n", book.Title, book.Author, book.Price, book.InStock)

		// Try selling a book
		if err := book.SellBook(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Sold one copy of %s. Remaining stock: %d\n", book.Title, book.InStock)
		}
	}

	// Search for books with "Go" pattern
	pattern := "Go"
	books := store.FindBooksByTitle(pattern)
	fmt.Printf("Found %d books while searching for patter '%s'\n", len(books), pattern)
	for _, book := range books {
		fmt.Printf("Found book: %s by %s - Price: $%.2f, In Stock: %d\n", book.Title, book.Author, book.Price, book.InStock)
	}
}
