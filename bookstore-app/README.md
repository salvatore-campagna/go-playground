# Bookstore Inventory Manager

Welcome to the Bookstore Inventory Manager exercise! This exercise will help you practice working with structs in Go. You’ll create and manage a simple inventory system for a bookstore, learning how to build, manipulate, and interact with Go structs.

## Instructions

Create a package named `bookstore`. Inside, define the following structs and methods to represent and manage books in the store.

### 1. `Book` Struct

Define a `Book` struct to represent a single book with these fields:
- `Title` (string) - The title of the book.
- `Author` (string) - The author of the book.
- `Price` (float64) - The price of the book.
- `InStock` (int) - The number of copies available in stock.

### 2. `NewBook` Function

Implement a function `NewBook` with parameters for the title, author, price, and stock quantity. This function should return an instance of `Book` initialized with the provided values.

### 3. `HasStock` Method

Implement a `HasStock` method for the `Book` struct. This method should return `true` if the `InStock` count is greater than zero, indicating the book is available.

### 4. `SellBook` Method

Implement a `SellBook` method for `Book`. This method should:
- Decrease the `InStock` count by 1 if the book is in stock.
- Return an error if `InStock` is zero, indicating that the book is out of stock.

### 5. `Bookstore` Struct

Define a `Bookstore` struct to represent the store itself. It should contain:
- `Name` (string) - The name of the bookstore.
- `Books` ([]Book) - A slice of `Book` structs representing the inventory.

### 6. `AddBook` Method

Implement an `AddBook` method for `Bookstore`. This method should add a `Book` to the `Books` slice.

### 7. `TotalInventoryValue` Method

Implement a `TotalInventoryValue` method for `Bookstore`. This method should calculate and return the total value of all books in stock by summing up `Price * InStock` for each book.

### 8. `FindBookByTitle` Method

Implement a `FindBookByTitle` method for `Bookstore`. This method should:
- Take a `title` string as an argument.
- Search for a `Book` with a matching title in the `Books` slice.
- Return the `Book` if found, or a default `Book` struct if it doesn’t exist.

### 9. `FindBooksByTitle` Method

Implement a `FindBooksByTitle` method for `Bookstore`. This method should:
- Take a `pattern` string as an argument.
- Search for books with titles containing the pattern as a substring in the `Books` slice.
- Return a slice of all matching `Book` structs.
