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

	// Term returns the term associated with this iterator.
	Term() string

	// TermFrequency returns the term frequency associated with the current document ID.
	TermFrequency() (float32, error)
}

// RoaringBitmapIterator implements BitmapIterator for iterating over a RoaringBitmap container.
type RoaringBitmapIterator struct {
	bitmap        *RoaringBitmap   // The RoaringBitmap being iterated
	keys          []uint16         // Keys identifying the containers in the bitmap
	currentKey    int              // Index of the current key
	container     RoaringContainer // Current container being iterated
	currentDocID  uint32           // Current document ID
	index         int              // Current index within the container
	term          string           // Term associated with this iterator
	termFrequency float32          // Mock or actual term frequency
}

// NewRoaringBitmapIterator creates a new iterator for a RoaringBitmap and its associated term.
func NewRoaringBitmapIterator(bitmap *RoaringBitmap, term string) *RoaringBitmapIterator {
	keys := make([]uint16, 0, len(bitmap.containers))
	for key := range bitmap.containers {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return &RoaringBitmapIterator{
		bitmap:        bitmap,
		keys:          keys,
		currentKey:    -1,
		container:     nil,
		currentDocID:  0,
		index:         -1,
		term:          term,
		termFrequency: 1.0, // Placeholder; replace with actual frequency logic if available
	}
}

// Next advances to the next document ID in the bitmap.
func (it *RoaringBitmapIterator) Next() (bool, error) {
	// Advance to the next valid document ID
	for {
		// Check if we need to move to the next container
		if it.container == nil || it.index >= it.container.Cardinality()-1 {
			it.currentKey++
			if it.currentKey >= len(it.keys) {
				// No more containers, iteration is complete
				return false, nil
			}
			// Update the container for the new key
			key := it.keys[it.currentKey]
			it.container = it.bitmap.containers[key]
			it.index = -1 // Reset index for the new container
		}

		// Advance within the current container
		it.index++
		if it.index < it.container.Cardinality() {
			if arrayContainer, ok := it.container.(*ArrayContainer); ok {
				it.currentDocID = uint32(it.keys[it.currentKey])<<16 | uint32(arrayContainer.values[it.index])
			} else if bitmapContainer, ok := it.container.(*BitmapContainer); ok {
				// Find the document ID for the current index
				count := 0
				for i, word := range bitmapContainer.Bitmap {
					for j := 0; j < 64; j++ {
						if word&(1<<j) != 0 {
							if count == it.index {
								it.currentDocID = uint32(it.keys[it.currentKey])<<16 | uint32(i*64+j)
								break
							}
							count++
						}
					}
				}
			}
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
		return key | uint32(arrayContainer.values[it.index]), nil
	}
	if bitmapContainer, ok := it.container.(*BitmapContainer); ok {
		count := 0
		for i, word := range bitmapContainer.Bitmap {
			for j := 0; j < 64; j++ {
				if word&(1<<j) != 0 {
					if count == it.index {
						return key | uint32(i*64+j), nil
					}
					count++
				}
			}
		}
	}
	return 0, fmt.Errorf("unknown container type")
}

// Term returns the term associated with the iterator.
func (it *RoaringBitmapIterator) Term() string {
	return it.term
}

// TermFrequency returns the term frequency associated with the current document ID.
func (it *RoaringBitmapIterator) TermFrequency() (float32, error) {
	// This is a placeholder. Implement actual logic to fetch term frequency from metadata if needed.
	return it.termFrequency, nil
}

// PostingListIterator defines an interface for iterating over posting lists.
// It provides methods to traverse document IDs and retrieve term frequencies.
type PostingListIterator interface {
	// Next advances the iterator to the next document ID in the posting list.
	Next() (bool, error)

	// DocID returns the current document ID in the posting list.
	DocID() (uint32, error)

	// Term returns the term associated with this iterator.
	Term() string

	// TermFrequency returns the term frequency associated with the current document ID.
	TermFrequency() (float32, error)

	// CurrentBlock returns the current block being processed by the iterator.
	CurrentBlock() *Block
}

// TermIterator implements PostingListIterator for traversing term posting lists in blocks.
type TermIterator struct {
	blocks        []*Block       // Posting list blocks for the term
	currentBlock  int            // Index of the current block
	blockIterator BitmapIterator // Iterator for the current block's bitmap
	currentDocID  uint32         // Current document ID
	term          string         // Term associated with this iterator
}

// NewTermIterator creates a new TermIterator for the given blocks.
func NewTermIterator(blocks []*Block, term string) PostingListIterator {
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
		blockIterator: firstBlock.Bitmap.BitmapIterator(),
		term:          term,
	}
}

// Next advances to the next document in the posting list.
func (it *TermIterator) Next() (bool, error) {
	for {
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
		it.blockIterator = it.blocks[it.currentBlock].Bitmap.BitmapIterator()
	}
}

// DocID retrieves the current document ID.
func (it *TermIterator) DocID() (uint32, error) {
	return it.currentDocID, nil
}

// Term retrieves the term associated with the iterator.
func (it *TermIterator) Term() string {
	return it.term
}

// TermFrequency retrieves the term frequency for the current document.
func (it *TermIterator) TermFrequency() (float32, error) {
	// Validate currentBlock is within range
	if it.currentBlock < 0 || it.currentBlock >= len(it.blocks) {
		return 0, fmt.Errorf("invalid block index %d while retrieving term frequency", it.currentBlock)
	}

	block := it.blocks[it.currentBlock]

	// Calculate rank
	rank, err := block.Bitmap.Rank(it.currentDocID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate rank for docID %d: %w", it.currentDocID, err)
	}

	// Ensure the rank is valid and map it to 0-based indexing
	if rank <= 0 || rank > len(block.TermFrequencies) {
		return 0, fmt.Errorf("rank %d out of bounds for term frequencies (len=%d)", rank, len(block.TermFrequencies))
	}

	return block.TermFrequencies[rank-1], nil
}

// CurrentBlock returns the current block being processed by the iterator.
func (it *TermIterator) CurrentBlock() *Block {
	if it.currentBlock >= 0 && it.currentBlock < len(it.blocks) {
		return it.blocks[it.currentBlock]
	}
	return nil
}

// TermIterator returns a PostingListIterator for the specified term.
func (s *Segment) TermIterator(term string) (PostingListIterator, error) {
	termMetadata, exists := s.Terms[term]
	if !exists {
		return &EmptyIterator{}, nil
	}
	return NewTermIterator(termMetadata.Blocks, term), nil
}

// TermIterators returns PostingListIterators for a list of terms.
func (s *Segment) TermIterators(terms []string) ([]PostingListIterator, error) {
	var termIterators []PostingListIterator
	for _, term := range terms {
		termIterator, err := s.TermIterator(term)
		if err != nil {
			return nil, err
		}
		termIterators = append(termIterators, termIterator)
	}

	return termIterators, nil
}

// BitmapIterator returns a BitmapIterator for the RoaringBitmap.
func (rb *RoaringBitmap) BitmapIterator() BitmapIterator {
	keys := make([]uint16, 0, len(rb.containers))
	for key := range rb.containers {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return &RoaringBitmapIterator{
		bitmap:     rb,
		keys:       keys,
		currentKey: -1,
	}
}

// EmptyIterator provides a no-op implementation of PostingListIterator.
type EmptyIterator struct{}

// Next always returns false, indicating there are no elements to iterate over.
func (it *EmptyIterator) Next() (bool, error) {
	return false, nil
}

// DocID returns an error because there are no valid elements in the iterator.
func (it *EmptyIterator) DocID() (uint32, error) {
	return 0, fmt.Errorf("no valid DocID in empty iterator")
}

// Term retrieves the term for an empty iterator (always empty string).
func (it *EmptyIterator) Term() string {
	return ""
}

// TermFrequency returns an error because there are no valid elements in the iterator.
func (it *EmptyIterator) TermFrequency() (float32, error) {
	return 0, fmt.Errorf("no valid TermFrequency in empty iterator")
}

// CurrentBlock returns nil because there are no blocks in an empty iterator.
func (it *EmptyIterator) CurrentBlock() *Block {
	return nil
}
