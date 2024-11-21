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
	container    RoaringContainer // Current container being iterated
	currentDocID uint32           // Current document ID
	index        int              // Current index within the container
}

func NewRoaringBitmapIterator(bitmap *RoaringBitmap) *RoaringBitmapIterator {
	keys := make([]uint16, 0, len(bitmap.containers))
	for key := range bitmap.containers {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] }) // Ensure keys are sorted

	return &RoaringBitmapIterator{
		bitmap:     bitmap,
		keys:       keys,
		currentKey: -1, // Start before the first key
		index:      -1, // Start before the first index
	}
}

// Next advances the iterator to the next document ID in the bitmap.
// It handles switching between containers and returns false if there are no more document IDs.
func (it *RoaringBitmapIterator) Next() (bool, error) {
	for {
		if it.currentKey == -1 {
			it.currentKey = 0
			if it.currentKey >= len(it.keys) {
				return false, nil
			}
			key := it.keys[it.currentKey]
			container, exists := it.bitmap.containers[key]
			if !exists {
				return false, fmt.Errorf("container not found for key %d", key)
			}
			it.container = container
			it.index = -1
		}

		if it.index >= -1 && it.index < it.container.Cardinality()-1 {
			it.index++
			docID, err := it.GetDocID()
			if err == nil {
				it.currentDocID = docID
				return true, nil
			}
			return false, err
		}

		it.currentKey++
		if it.currentKey >= len(it.keys) {
			return false, nil
		}

		key := it.keys[it.currentKey]
		container, exists := it.bitmap.containers[key]
		if !exists {
			return false, fmt.Errorf("container not found for key %d", key)
		}
		it.container = container
		it.index = -1
	}
}

// GetDocID calculates the document ID for the current index within the container.
func (it *RoaringBitmapIterator) GetDocID() (uint32, error) {
	key := uint32(it.keys[it.currentKey]) << 16
	if arrayContainer, ok := it.container.(*ArrayContainer); ok {
		if it.index < len(arrayContainer.values) {
			return key | uint32(arrayContainer.values[it.index]), nil
		}
		return 0, fmt.Errorf("index out of bounds in array container")
	} else if bitmapContainer, ok := it.container.(*BitmapContainer); ok {
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
		return 0, fmt.Errorf("index out of bounds in bitmap container")
	}
	return 0, fmt.Errorf("unknown container type")
}

// DocID returns the current document ID of the iterator.
func (it *RoaringBitmapIterator) DocID() (uint32, error) {
	if it.currentKey < 0 || it.currentKey >= len(it.keys) {
		return 0, fmt.Errorf("current key %d is out of bounds", it.currentKey)
	}
	return it.currentDocID, nil
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
	blocks          []*Block       // Posting list blocks for the term
	currentBlock    int            // Index of the current block
	bitmapIterator  BitmapIterator // Iterator for the current block's bitmap
	currentDocId    uint32         // Current document ID
	currentTermFreq float32        // Current term frequency
	blockIndex      int            // Current index within the block
}

// Next advances the iterator to the next document in the posting list.
// It switches blocks when necessary and calculates term frequencies for documents.
func (it *TermIterator) Next() (bool, error) {
	for {
		if it.bitmapIterator != nil {
			hasNext, err := it.bitmapIterator.Next()
			if err != nil {
				return false, err
			}

			if hasNext {
				it.currentDocId, err = it.bitmapIterator.DocID()
				if err != nil {
					return false, err
				}

				it.currentTermFreq, err = it.getTermFrequency(it.currentDocId)
				if err != nil {
					return false, err
				}
				return true, nil
			}
		}

		it.currentBlock++
		if it.currentBlock >= len(it.blocks) {
			return false, nil
		}

		it.bitmapIterator = it.blocks[it.currentBlock].Bitmap.Iterator()
	}
}

// DocID retrieves the current document ID in the posting list.
func (it *TermIterator) DocID() (uint32, error) {
	if it.bitmapIterator == nil {
		return 0, fmt.Errorf("bitmap iterator is null")
	}
	return it.currentDocId, nil
}

// TermFrequency retrieves the term frequency for the current document ID.
func (it *TermIterator) TermFrequency() (float32, error) {
	if it.bitmapIterator == nil {
		return 0, fmt.Errorf("bitmap iterator is null")
	}
	return it.currentTermFreq, nil
}

// getTermFrequency retrieves the term frequency for the specified document ID.
func (it *TermIterator) getTermFrequency(docID uint32) (float32, error) {
	if it.currentBlock < 0 || it.currentBlock >= len(it.blocks) {
		return 0, fmt.Errorf("block index %d is out of bounds", it.currentBlock)
	}

	block := it.blocks[it.currentBlock]
	if !block.Bitmap.Contains(docID) {
		return 0, fmt.Errorf("docID %d does not exist in the bitmap", docID)
	}

	rank, err := block.Bitmap.Rank(docID)
	if err != nil {
		return 0, err
	}

	if rank < 0 || rank > len(block.TermFrequencies) {
		return 0, fmt.Errorf("rank %d is out of bounds for term frequencies with len %d", rank, len(block.TermFrequencies))
	}
	return block.TermFrequencies[rank-1], nil
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
