// Package chessboard provides functions for counting occupied squares
// on a chessboard represented using maps and slices.
package chessboard

// File represents a column (file) on the chessboard. It stores whether each square in the file is occupied.
type File []bool

// Chessboard represents an 8x8 chessboard using a map where keys are file names ("A" to "H")
// and values are File slices indicating which squares are occupied.
type Chessboard map[string]File

// CountInFile returns the number of occupied squares in the specified file.
// If the file does not exist, it returns 0.
func CountInFile(chessboard Chessboard, file string) int {
	fileCounter := 0
	for _, square := range chessboard[file] {
		if square {
			fileCounter++
		}
	}
	return fileCounter
}

// CountInRank returns the number of occupied squares in the specified rank (row).
// If the rank is out of bounds (less than 1 or greater than 8), it returns 0.
func CountInRank(chessboard Chessboard, rank int) int {
	if rank < 1 || rank > 8 {
		return 0
	}

	rankCounter := 0
	for _, file := range chessboard {
		if rank <= len(file) && file[rank-1] {
			rankCounter++
		}
	}
	return rankCounter
}

// CountAll returns the total number of squares present in the chessboard.
// This function counts all squares, even if some files have fewer than 8 squares.
func CountAll(chessboard Chessboard) int {
	total := 0
	for _, file := range chessboard {
		total += len(file)
	}
	return total
}

// CountOccupied returns the total number of occupied squares on the chessboard.
func CountOccupied(chessboard Chessboard) int {
	total := 0
	for file := range chessboard {
		total += CountInFile(chessboard, file)
	}
	return total
}
