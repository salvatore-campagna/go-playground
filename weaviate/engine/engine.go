// Package engine provides a query execution engine for full-text search over inverted index segments.
// It supports multi-term queries and ranking of documents based on relevance scores. The engine is designed
// for efficient traversal of posting lists, leveraging heap-based priority queues for block processing
// and TF-IDF scoring for relevance computation.
//
// # Features
//
// - Supports multi-term queries across multiple segments.
// - Efficient block-based processing using min-heaps for priority management.
// - TF-IDF scoring for relevance computation, ensuring accurate ranking of results.
// - Supports extension with custom ranking functions.
//
// # TODOs
//
// - Parallelize query execution for better performance on multi-core systems.

package engine

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"weaviate/storage"
)

// ScoredDocument represents a document with its associated score.
type ScoredDocument struct {
	DocID uint32
	Score float64
}

// QueryEngine defines the interface for executing queries on an index.
type QueryEngine interface {
	// MultiTermQuery performs a query on multiple terms and ranks the results using the provided comparator.
	// The comparator is a function that determines the ranking order of the scored documents.
	MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error)
}

// queryEngine is the default implementation of the QueryEngine interface.
type queryEngine struct {
	segments  []*storage.Segment // List of index segments to query.
	totalDocs uint32             // Total number of documents across all segments.
}

// NewQueryEngine initializes a new QueryEngine with the given segments and total document count.
// Returns an error if the input parameters are invalid.
func NewQueryEngine(segments []*storage.Segment, totalDocs uint32) (QueryEngine, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments to query")
	}
	if totalDocs == 0 {
		return nil, fmt.Errorf("totalDocs must be greater than zero")
	}
	return &queryEngine{
		segments:  segments,
		totalDocs: totalDocs,
	}, nil
}

// blockEntry represents an entry in the min-heap for block processing.
type blockEntry struct {
	block    *storage.Block         // The block being processed.
	iterator storage.BitmapIterator // Iterator over the block's bitmap.
	docID    uint32                 // The current document ID.
}

// minBlockHeap is a priority queue (min-heap) for managing block entries during query execution.
type minBlockHeap []*blockEntry

// Len returns the number of elements in the heap.
func (h minBlockHeap) Len() int { return len(h) }

// Less determines the order of elements in the heap based on their MinDocID.
func (h minBlockHeap) Less(i, j int) bool {
	return h[i].block.MinDocID < h[j].block.MinDocID
}

// Swap exchanges two elements in the heap.
func (h minBlockHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push adds an element to the heap.
func (h *minBlockHeap) Push(x interface{}) {
	*h = append(*h, x.(*blockEntry))
}

// Pop removes and returns the smallest element from the heap.
func (h *minBlockHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// Top returns the top element of the min-heap without removing it.
func (h *minBlockHeap) Top() *blockEntry {
	if len(*h) > 0 {
		return (*h)[0]
	}
	return nil
}

// getTermPostingListIterators retrieves posting list iterators for the given terms from all segments.
// Returns a map of terms to their respective iterators or an error if no iterators are found.
func getTermPostingListIterators(terms []string, segments []*storage.Segment) (map[string][]storage.PostingListIterator, error) {
	termIterators := make(map[string][]storage.PostingListIterator)

	for _, term := range terms {
		var iterators []storage.PostingListIterator
		for _, segment := range segments {
			iterator, err := segment.TermIterator(term)
			if err != nil {
				return nil, fmt.Errorf("error creating iterator for term %s: %v", term, err)
			}
			if _, ok := iterator.(*storage.EmptyIterator); !ok {
				iterators = append(iterators, iterator)
			}
		}
		if len(iterators) > 0 {
			termIterators[term] = iterators
		} else {
			return nil, fmt.Errorf("term %s not found in any segment", term)
		}
	}
	return termIterators, nil
}

func (qe *queryEngine) MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error) {
	// Step 1: Retrieve posting list iterators for each term in the query across all segments.
	termIterators, err := getTermPostingListIterators(terms, qe.segments)
	if err != nil {
		return nil, err
	}

	// Step 2: Initialize a min-heap to process posting list blocks efficiently.
	blockHeap := &minBlockHeap{}
	heap.Init(blockHeap)

	// Step 3: Push posting list blocks into the heap.
	for term, iterators := range termIterators {
		for _, iter := range iterators {
			hasNext, err := iter.Next()
			if err != nil {
				return nil, fmt.Errorf("error advancing iterator for term %s: %v", term, err)
			}
			if hasNext {
				docID, err := iter.DocID()
				if err != nil {
					return nil, fmt.Errorf("error retrieving DocID for term %s: %v", term, err)
				}
				block := iter.CurrentBlock()
				heap.Push(blockHeap, &blockEntry{
					block:    block,
					iterator: iter,
					docID:    docID,
				})
			}
		}
	}

	// Step 4: Initialize an array to store scored documents.
	var scoredDocuments []ScoredDocument

	// Step 5: Process the min-heap to evaluate documents.
	for blockHeap.Len() > 0 {
		// Check the smallest `docID` in the heap.
		top := (*blockHeap)[0]
		currentDocID := top.docID

		// Collect all entries matching the current `docID`.
		matchingEntries := []*blockEntry{}
		for _, entry := range *blockHeap {
			if entry.docID == currentDocID {
				matchingEntries = append(matchingEntries, entry)
			}
		}

		// Check if all terms are present in the matching entries.
		if len(matchingEntries) == len(terms) {
			// Compute the score for the matching document only if all terms match.
			var score float64
			for _, entry := range matchingEntries {
				// Retrieve the term frequency
				termFrequency, err := entry.iterator.TermFrequency()
				if err != nil {
					return nil, fmt.Errorf("error retrieving TermFrequency: %v", err)
				}
				term := entry.iterator.Term()

				// Calculate document frequency (DF) for the term.
				df := 0
				for _, segment := range qe.segments {
					if metadata, exists := segment.Terms[term]; exists {
						df += int(metadata.TotalDocs)
					}
				}
				score += float64(termFrequency) * math.Log(float64(qe.totalDocs+1)/float64(df+1))
			}

			// Add the DocID and the corresponding score to the query result for the matched document.
			scoredDocuments = append(scoredDocuments, ScoredDocument{
				DocID: currentDocID,
				Score: score,
			})

			// Advance all iterators pointing to the currentDocID and matching the term.
			for _, entry := range matchingEntries {
				// Attempt to move the iterator to the next document in the posting list.
				hasNext, err := entry.iterator.Next()
				if err != nil {
					// If advancing the iterator fails, terminate the query execution and return the error.
					return nil, fmt.Errorf("error advancing iterator: %v", err)
				}

				if hasNext {
					// The iterator successfully advanced to the next document.
					// Retrieve the new `docID` from the iterator and update the heap entry.
					entry.docID, _ = entry.iterator.DocID()

					// Reorganize the heap to maintain the min-heap property.
					// This is necessary because the updated `docID` could affect the order of the heap.
					// Since we are always working with the top element, the index is known to be `0`.
					heap.Fix(blockHeap, 0)
				} else {
					// The iterator has been exhausted (no more documents in the posting list).
					// Remove the corresponding entry from the heap, as it no longer contributes
					// to the query results. The entry being processed is always at the top of the heap.
					heap.Remove(blockHeap, 0)
				}
			}
		} else {
			// Advance the iterator for the smallest docID.
			smallest := heap.Pop(blockHeap).(*blockEntry)
			hasNext, err := smallest.iterator.Next()
			if err != nil {
				return nil, fmt.Errorf("error advancing iterator: %v", err)
			}
			if hasNext {
				docID, err := smallest.iterator.DocID()
				if err != nil {
					return nil, fmt.Errorf("error retrieving next DocID: %v", err)
				}
				smallest.docID = docID
				heap.Push(blockHeap, smallest)
			}
		}
	}

	// Step 6: Sort the results based on the provided comparison function.
	sort.Slice(scoredDocuments, func(i, j int) bool {
		return less(scoredDocuments[i], scoredDocuments[j])
	})
	return scoredDocuments, nil
}
