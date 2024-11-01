// Package bookstore provides structures and methods for managing a bookstore's inventory,
// allowing creation of books, addition to the store, and various operations to check
// inventory, sell books, and calculate the total value of available stock.
package bookstore

import (
	"fmt"
	"strings"
)

// Book represents a book with details including title, author, price, and quantity in stock.
type Book struct {
	Title   string  // Title is the name of the book.
	Author  string  // Author is the person who wrote the book.
	Price   float64 // Price is the cost of one copy of the book.
	InStock int     // InStock indicates the number of copies available in the store.
}

// BookStore represents a bookstore with a name and a collection of books.
type BookStore struct {
	Name  string // Name is the name of the bookstore.
	Books []Book // Books holds the collection of books available in the bookstore.
}

// NewBook creates and returns a new Book with the specified title, author, price, and initial stock.
func NewBook(title string, author string, price float64, inStock int) Book {
	return Book{
		Title:   title,
		Author:  author,
		Price:   price,
		InStock: inStock,
	}
}

// HasStock checks if there is at least one copy of the book in stock.
// It returns true if InStock is greater than zero, otherwise false.
func (b *Book) HasStock() bool {
	return b.InStock > 0
}

// SellBook decrements the stock of the book by one if it is available.
// If the book is out of stock, it returns an error indicating that the book cannot be sold.
func (b *Book) SellBook() error {
	if b.InStock < 1 {
		return fmt.Errorf("book with title '%s' not found", b.Title)
	}
	b.InStock--
	return nil
}

// NewBookStore creates a new BookStore with the given name and initializes an empty book inventory.
func NewBookStore(name string) BookStore {
	return BookStore{
		Name:  name,
		Books: make([]Book, 0),
	}
}

// AddBook adds a specified Book to the bookstore's inventory.
func (bs *BookStore) AddBook(b Book) {
	bs.Books = append(bs.Books, b)
}

// TotalInventoryValue calculates and returns the total value of all books currently in stock,
// based on the price and quantity of each book.
func (bs *BookStore) TotalInventoryValue() float64 {
	totalInventoryValue := 0.0
	for _, book := range bs.Books {
		totalInventoryValue += (book.Price * float64(book.InStock))
	}
	return totalInventoryValue
}

// FindBookByTitle searches the bookstore's inventory for a book with a title
// that exactly matches the specified title. It returns the book if found,
// or an error indicating that no matching book was found.
func (bs *BookStore) FindBookByTitle(title string) (Book, error) {
	for _, book := range bs.Books {
		if book.Title == title {
			return book, nil
		}
	}
	return Book{}, fmt.Errorf("book with title '%s' not found", title)
}

// FindBooksByTitle searches the bookstore's inventory for books with titles that
// contain the specified pattern as a substring. It returns a slice of matching books.
func (bs *BookStore) FindBooksByTitle(pattern string) []Book {
	books := make([]Book, 0)
	for _, book := range bs.Books {
		if strings.Contains(book.Title, pattern) {
			books = append(books, book)
		}
	}
	return books
}
