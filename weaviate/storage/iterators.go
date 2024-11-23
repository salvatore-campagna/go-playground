// Package storage provides data structures and iterators for efficient storage and retrieval
// of posting lists and term frequencies using Roaring Bitmaps. This package enables efficient
// queries and data traversal in search engines or inverted index implementations.
package storage

import (
	"fmt"
	"sort"
)

// BitmapIterator defines an interface for iterating over document IDs stored in a bitmap.
type BitmapIterator interface {
	// Next advances the iterator to the next document ID. It returns true if there is a next document ID,
	// false otherwise. Any error encountered during iteration is returned.
	Next() (bool, error)

	// DocID returns the current document ID pointed to by the iterator. If no valid document is available,
	// it returns an error.
	DocID() (uint32, error)
}

// RoaringBitmapIterator implements BitmapIterator for iterating over a RoaringBitmap container.
type RoaringBitmapIterator struct {
	bitmap       *RoaringBitmap   // The RoaringBitmap being iterated
	keys         []uint16         // Keys identifying the containers in the bitmap
	currentKey   int              // Index of the current key
	currentIndex int              // CUrrent position in the container
	container    RoaringContainer // Current container being iterated
}

func NewRoaringBitmapIterator(bitmap *RoaringBitmap) *RoaringBitmapIterator {
	keys := make([]uint16, 0, len(bitmap.containers))
	for key := range bitmap.containers {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] }) // Ensure keys are sorted

	return &RoaringBitmapIterator{
		bitmap:       bitmap,
		keys:         keys,
		currentKey:   -1,
		currentIndex: -1,
	}
}

// Next advances to the next document ID.
func (it *RoaringBitmapIterator) Next() (bool, error) {
	for {
		// Move to the next container if needed
		if it.currentKey == -1 || it.currentIndex >= it.container.Cardinality()-1 {
			it.currentKey++
			if it.currentKey >= len(it.keys) {
				return false, nil // No more containers
			}
			key := it.keys[it.currentKey]
			it.container = it.bitmap.containers[key]
			it.currentIndex = -1
		}

		// Advance within the current container
		it.currentIndex++
		if it.currentIndex < it.container.Cardinality() {
			return true, nil
		}
	}
}

// DocID retrieves the current document ID.
func (it *RoaringBitmapIterator) DocID() (uint32, error) {
	if it.currentKey < 0 || it.currentKey >= len(it.keys) {
		return 0, fmt.Errorf("invalid key while iterating container")
	}
	key := uint32(it.keys[it.currentKey]) << 16
	if arrayContainer, ok := it.container.(*ArrayContainer); ok {
		return key | uint32(arrayContainer.values[it.currentIndex]), nil
	}
	if bitmapContainer, ok := it.container.(*BitmapContainer); ok {
		count := 0
		for i, word := range bitmapContainer.Bitmap {
			for j := 0; j < 64; j++ {
				if word&(1<<j) != 0 {
					if count == it.currentIndex {
						return key | uint32(i*64+j), nil
					}
					count++
				}
			}
		}
	}
	return 0, fmt.Errorf("unknown container type")
}

// PostingListIterator defines an interface for iterating over posting lists.
// It provides methods to traverse document IDs and retrieve term frequencies.
type PostingListIterator interface {
	// Next advances the iterator to the next document ID in the posting list.
	Next() (bool, error)

	// DocID returns the current document ID in the posting list.
	DocID() (uint32, error)

	// TermFrequency retrieves the term frequency for the current document ID.
	TermFrequency() (float32, error)
}

// TermIterator implements PostingListIterator for traversing term posting lists in blocks.
type TermIterator struct {
	blocks        []*Block       // Posting list blocks for the term
	currentBlock  int            // Index of the current block
	blockIterator BitmapIterator // Iterator for the current block's bitmap
	currentDocID  uint32         // Current document ID
}

// NewTermIterator creates a new TermIterator for the given blocks.
func NewTermIterator(blocks []*Block) PostingListIterator {
	if len(blocks) == 0 {
		return &EmptyIterator{}
	}

	firstBlock := blocks[0]
	if firstBlock == nil || firstBlock.Bitmap == nil {
		return &EmptyIterator{}
	}

	return &TermIterator{
		blocks:        blocks,
		currentBlock:  0,
		blockIterator: firstBlock.Bitmap.Iterator(),
	}
}

// Next advances to the next document in the posting list.
func (it *TermIterator) Next() (bool, error) {
	for {
		// Try advancing within the current block
		if it.blockIterator != nil {
			hasNext, err := it.blockIterator.Next()
			if err != nil {
				return false, err
			}
			if hasNext {
				docID, err := it.blockIterator.DocID()
				if err != nil {
					return false, err
				}
				it.currentDocID = docID
				return true, nil
			}
		}

		// Move to the next block
		it.currentBlock++
		if it.currentBlock >= len(it.blocks) {
			return false, nil // No more blocks
		}
		it.blockIterator = it.blocks[it.currentBlock].Bitmap.Iterator()
	}
}

// DocID retrieves the current document ID.
func (it *TermIterator) DocID() (uint32, error) {
	return it.currentDocID, nil
}

// TermFrequency retrieves the term frequency for the current document.
func (it *TermIterator) TermFrequency() (float32, error) {
	block := it.blocks[it.currentBlock]
	rank, err := block.Bitmap.Rank(it.currentDocID)
	if err != nil {
		return 0, err
	}
	if rank <= 0 || rank > len(block.TermFrequencies) {
		return 0, fmt.Errorf("rank out of bounds for term frequencies")
	}
	return block.TermFrequencies[rank-1], nil
}

// CoordinatedIterator handles intersection of multiple term iterators.
type CoordinatedIterator struct {
	iterators []PostingListIterator // List of term iterators
}

// NewCoordinatedIterator creates a new iterator for multiple terms.
func NewCoordinatedIterator(iterators []PostingListIterator) *CoordinatedIterator {
	return &CoordinatedIterator{iterators: iterators}
}

// Next advances to the next document common to all terms.
func (ci *CoordinatedIterator) Next() (bool, error) {
	for {
		minDocID := uint32(0)
		allMatch := true

		// Find the smallest doc ID among iterators
		for _, it := range ci.iterators {
			docID, err := it.DocID()
			if err != nil {
				return false, err
			}
			if docID > minDocID {
				minDocID = docID
				allMatch = false
			}
		}

		// If all iterators match, return true
		if allMatch {
			return true, nil
		}

		// Advance iterators that are behind
		for _, it := range ci.iterators {
			docID, err := it.DocID()
			if err != nil {
				return false, err
			}
			for docID < minDocID {
				hasNext, err := it.Next()
				if err != nil || !hasNext {
					return false, err
				}
				docID, _ = it.DocID()
			}
		}
	}
}

// DocID retrieves the current document ID.
func (ci *CoordinatedIterator) DocID() (uint32, error) {
	if len(ci.iterators) == 0 {
		return 0, fmt.Errorf("no iterators available")
	}
	return ci.iterators[0].DocID()
}

// EmptyIterator provides a no-op implementation of PostingListIterator.
type EmptyIterator struct{}

func (it *EmptyIterator) Next() (bool, error) {
	return false, fmt.Errorf("next on empty iterator with no elements")
}

func (it *EmptyIterator) DocID() (uint32, error) {
	return 0, fmt.Errorf("doc id on empty empty iterator with no elements")
}

func (it *EmptyIterator) TermFrequency() (float32, error) {
	return 0, fmt.Errorf("term frequency on empty iterator with no elements")
}
