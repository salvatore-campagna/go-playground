package storage

// # TODOs
//
//   - Add support for filtering documents during iteration.
//   - Add more set operations (e.g., difference, XOR) to iterators for advanced queries.
//   - Introduce custom error types for iterator-related errors.
//   - Add checksums to ensure data consistency during iteration and storage.

import (
	"fmt"
	"sort"
)

// BitmapIterator defines an interface for iterating over document IDs stored in a bitmap.
type BitmapIterator interface {
	Next() (bool, error)
	DocID() (uint32, error)
	Term() string
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
	termFrequency float32          // Term frequency of the term associated with this iterator
}

// NewRoaringBitmapIterator creates a new iterator for a RoaringBitmap and its associated term.
func NewRoaringBitmapIterator(bitmap *RoaringBitmap, term string, termFrequency float32) *RoaringBitmapIterator {
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
		termFrequency: termFrequency,
	}
}

// Next advances to the next document ID in the bitmap.
func (it *RoaringBitmapIterator) Next() (bool, error) {
	for {
		// Move to the next container if no container or end of current container
		if it.container == nil || it.index >= it.container.Cardinality() {
			it.currentKey++
			if it.currentKey >= len(it.keys) {
				// No more containers, iteration is complete
				return false, nil
			}

			// Move to the next container
			key := it.keys[it.currentKey]
			it.container = it.bitmap.containers[key]

			// Skip empty containers
			if it.container.Cardinality() == 0 {
				continue
			}

			// Reset index for the new container
			it.index = 0
		}

		// Move inside the current container
		if arrayContainer, ok := it.container.(*ArrayContainer); ok {
			it.currentDocID = uint32(it.keys[it.currentKey])<<16 | uint32(arrayContainer.values[it.index])
			it.index++
			return true, nil
		} else if bitmapContainer, ok := it.container.(*BitmapContainer); ok {
			// Find the next set bit in the bitmap
			for i := it.index / 64; i < len(bitmapContainer.bitmap); i++ {
				word := bitmapContainer.bitmap[i]
				for j := 0; j < 64; j++ {
					if word&(1<<j) != 0 {
						if (i*64 + j) >= it.index {
							it.currentDocID = uint32(it.keys[it.currentKey])<<16 | uint32(i*64+j)
							it.index = i*64 + j + 1
							return true, nil
						}
					}
				}
			}
			// If no more bits are set, move to the next container
			it.index = bitmapContainer.Cardinality() // Exhaust current container
		}
	}
}

// DocID retrieves the current document ID.
func (it *RoaringBitmapIterator) DocID() (uint32, error) {
	if it.currentKey < 0 || it.currentKey >= len(it.keys) {
		return 0, fmt.Errorf("invalid key while iterating container")
	}

	return it.currentDocID, nil
}

// Term returns the term associated with the iterator.
func (it *RoaringBitmapIterator) Term() string {
	return it.term
}

// TermFrequency returns the term frequency associated with the current document ID.
func (it *RoaringBitmapIterator) TermFrequency() (float32, error) {
	return it.termFrequency, nil
}

// PostingListIterator defines an interface for iterating over posting lists.
// It provides methods to traverse document IDs and retrieve term frequencies.
type PostingListIterator interface {
	Next() (bool, error)
	DocID() (uint32, error)
	Term() string
	TermFrequency() (float32, error)
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
	if firstBlock == nil || firstBlock.Bitmap == nil || firstBlock.Bitmap.Cardinality() == 0 {
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
			return false, nil
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
	if it.currentBlock < 0 || it.currentBlock >= len(it.blocks) {
		return 0, fmt.Errorf("invalid block index %d while retrieving term frequency", it.currentBlock)
	}

	block := it.blocks[it.currentBlock]

	// rank is the index (+1) used to access Block.TermFrequencies
	rank, err := block.Bitmap.Rank(it.currentDocID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate rank for docID %d: %w", it.currentDocID, err)
	}
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
	return 0, fmt.Errorf("invalid DocID in empty iterator")
}

// Term retrieves the term for an empty iterator (always empty string).
func (it *EmptyIterator) Term() string {
	return ""
}

// TermFrequency returns an error because there are no valid elements in the iterator.
func (it *EmptyIterator) TermFrequency() (float32, error) {
	return 0, fmt.Errorf("invalid term frequency in empty iterator")
}

// CurrentBlock returns nil because there are no blocks in an empty iterator.
func (it *EmptyIterator) CurrentBlock() *Block {
	return nil
}
