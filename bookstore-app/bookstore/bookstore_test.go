package bookstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBook(t *testing.T) {
	book := NewBook("Go in Action", "William Kennedy", 39.99, 10)
	assert.Equal(t, "Go in Action", book.Title, "Title should match")
	assert.Equal(t, "William Kennedy", book.Author, "Author should match")
	assert.Equal(t, 39.99, book.Price, "Price should match")
	assert.Equal(t, 10, book.InStock, "InStock should match")
}

func TestHasStock(t *testing.T) {
	book := NewBook("Go in Action", "William Kennedy", 39.99, 1)
	assert.True(t, book.HasStock(), "HasStock should return true when InStock > 0")
	book.InStock = 0
	assert.False(t, book.HasStock(), "HasStock should return false when InStock == 0")
}

func TestSellBook(t *testing.T) {
	book := NewBook("Go in Action", "William Kennedy", 39.99, 1)
	err := book.SellBook()
	assert.NoError(t, err, "Selling a book in stock should not return an error")
	assert.Equal(t, 0, book.InStock, "Remaining stock should be 0 after sale")
	err = book.SellBook()
	assert.Error(t, err, "Selling out-of-stock book should return an error")
}

func TestNewBookStore(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	assert.Equal(t, "The Go Bookstore", store.Name, "Store name should match")
	assert.Empty(t, store.Books, "New store should have an empty book slice")
}

func TestAddBook(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	book := NewBook("Go in Action", "William Kennedy", 39.99, 10)
	store.AddBook(book)
	assert.Len(t, store.Books, 1, "Store should contain 1 book after addition")
	assert.Equal(t, "Go in Action", store.Books[0].Title, "Added book title should match")
}

func TestTotalInventoryValue(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	store.AddBook(NewBook("Go in Action", "William Kennedy", 39.99, 10))
	store.AddBook(NewBook("Learning Go", "Jon Bodner", 29.99, 5))
	expectedValue := (39.99 * 10) + (29.99 * 5)
	assert.Equal(t, expectedValue, store.TotalInventoryValue(), "Total inventory value should match expected")
}

func TestFindBookByTitle(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	book := NewBook("Go in Action", "William Kennedy", 39.99, 10)
	store.AddBook(book)

	foundBook, err := store.FindBookByTitle("Go in Action")
	assert.NoError(t, err, "Finding an existing book should not return an error")
	assert.Equal(t, "Go in Action", foundBook.Title, "Found book title should match")

	_, err = store.FindBookByTitle("Unknown Book")
	assert.Error(t, err, "Finding a non-existing book should return an error")
}

func TestFindBooksByTitle(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	store.AddBook(NewBook("Go in Action", "William Kennedy", 39.99, 10))
	store.AddBook(NewBook("Go Fundamentals", "Various Authors", 29.99, 5))
	store.AddBook(NewBook("Python Basics", "Someone Else", 25.99, 7))

	results := store.FindBooksByTitle("Go")
	assert.Len(t, results, 2, "Expected 2 books to match the title pattern 'Go'")
	for _, book := range results {
		assert.Contains(t, book.Title, "Go", "Book title should contain 'Go'")
	}
}
