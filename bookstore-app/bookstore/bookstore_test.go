package bookstore

import (
	"strings"
	"testing"
)

func TestNewBook(t *testing.T) {
	book := NewBook("Go in Action", "William Kennedy", 39.99, 10)
	if book.Title != "Go in Action" {
		t.Errorf("expected title to be 'Go in Action', got %s", book.Title)
	}
	if book.Author != "William Kennedy" {
		t.Errorf("expected author to be 'William Kennedy', got %s", book.Author)
	}
	if book.Price != 39.99 {
		t.Errorf("expected price to be 39.99, got %f", book.Price)
	}
	if book.InStock != 10 {
		t.Errorf("expected inStock to be 10, got %d", book.InStock)
	}
}

func TestHasStock(t *testing.T) {
	book := NewBook("Go in Action", "William Kennedy", 39.99, 1)
	if !book.HasStock() {
		t.Errorf("expected HasStock to return true, got false")
	}
	book.InStock = 0
	if book.HasStock() {
		t.Errorf("expected HasStock to return false, got true")
	}
}

func TestSellBook(t *testing.T) {
	book := NewBook("Go in Action", "William Kennedy", 39.99, 1)
	if err := book.SellBook(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if book.InStock != 0 {
		t.Errorf("expected inStock to be 0, got %d", book.InStock)
	}
	err := book.SellBook()
	if err == nil {
		t.Errorf("expected an error when selling out-of-stock book, got nil")
	}
}

func TestNewBookStore(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	if store.Name != "The Go Bookstore" {
		t.Errorf("expected name to be 'The Go Bookstore', got %s", store.Name)
	}
	if len(store.Books) != 0 {
		t.Errorf("expected no books in the new store, got %d", len(store.Books))
	}
}

func TestAddBook(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	book := NewBook("Go in Action", "William Kennedy", 39.99, 10)
	store.AddBook(book)
	if len(store.Books) != 1 {
		t.Errorf("expected 1 book in the store, got %d", len(store.Books))
	}
	if store.Books[0].Title != "Go in Action" {
		t.Errorf("expected book title to be 'Go in Action', got %s", store.Books[0].Title)
	}
}

func TestTotalInventoryValue(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	store.AddBook(NewBook("Go in Action", "William Kennedy", 39.99, 10))
	store.AddBook(NewBook("Learning Go", "Jon Bodner", 29.99, 5))
	expectedValue := (39.99 * 10) + (29.99 * 5)
	if store.TotalInventoryValue() != expectedValue {
		t.Errorf("expected total inventory value to be %f, got %f", expectedValue, store.TotalInventoryValue())
	}
}

func TestFindBookByTitle(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	book := NewBook("Go in Action", "William Kennedy", 39.99, 10)
	store.AddBook(book)

	foundBook, err := store.FindBookByTitle("Go in Action")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if foundBook.Title != "Go in Action" {
		t.Errorf("expected book title to be 'Go in Action', got %s", foundBook.Title)
	}

	_, err = store.FindBookByTitle("Unknown Book")
	if err == nil {
		t.Errorf("expected an error when book is not found, got nil")
	}
}

func TestFindBooksByTitle(t *testing.T) {
	store := NewBookStore("The Go Bookstore")
	store.AddBook(NewBook("Go in Action", "William Kennedy", 39.99, 10))
	store.AddBook(NewBook("Go Fundamentals", "Various Authors", 29.99, 5))
	store.AddBook(NewBook("Python Basics", "Someone Else", 25.99, 7))

	results := store.FindBooksByTitle("Go")
	if len(results) != 2 {
		t.Errorf("expected 2 books to match the title pattern, got %d", len(results))
	}
	for _, book := range results {
		if !strings.Contains(book.Title, "Go") {
			t.Errorf("expected book title to contain 'Go', got %s", book.Title)
		}
	}
}
