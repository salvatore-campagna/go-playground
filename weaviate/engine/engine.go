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

// MultiTermQuery processes a query with multiple terms and returns the ranked results.
// The `less` function determines the ranking order of scored documents.
func (qe *queryEngine) MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error) {
	termIterators, err := getTermPostingListIterators(terms, qe.segments)
	if err != nil {
		return nil, err
	}

	blockHeap := &minBlockHeap{}
	heap.Init(blockHeap)

	termDF := make(map[string]int) // Document frequency (DF) for each term
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

		if len(iterators) > 0 {
			segment := qe.segments[0]
			if metadata, exists := segment.Terms[term]; exists {
				termDF[term] = int(metadata.TotalDocs)
			} else {
				termDF[term] = 0
			}
		}
	}

	scoredDocs := make(map[uint32]float64)
	docTermCounts := make(map[uint32]int)
	termDocTFs := make(map[string]map[uint32]float32)

	for _, term := range terms {
		termDocTFs[term] = make(map[uint32]float32)
	}

	for blockHeap.Len() > 0 {
		entry := heap.Pop(blockHeap).(*blockEntry)
		docID := entry.docID
		docTermCounts[docID]++
		tf, err := entry.iterator.TermFrequency()
		if err != nil {
			return nil, fmt.Errorf("error retrieving TermFrequency: %v", err)
		}
		termDocTFs[entry.iterator.Term()][docID] = tf

		if docTermCounts[docID] == len(terms) {
			totalScore := 0.0
			for term, termFrequencyMap := range termDocTFs {
				tf := termFrequencyMap[docID]
				df := termDF[term]
				idf := math.Log(float64(qe.totalDocs+1) / float64(df+1))
				tfidf := float64(tf) * idf
				totalScore += tfidf
			}
			scoredDocs[docID] = totalScore
		}

		hasNext, err := entry.iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error advancing iterator: %v", err)
		}
		if hasNext {
			docID, err := entry.iterator.DocID()
			if err != nil {
				return nil, fmt.Errorf("error retrieving next DocID: %v", err)
			}
			block := entry.block
			heap.Push(blockHeap, &blockEntry{
				block:    block,
				iterator: entry.iterator,
				docID:    docID,
			})
		}
	}

	var sortedScoredDocs []ScoredDocument
	for docID, score := range scoredDocs {
		sortedScoredDocs = append(sortedScoredDocs, ScoredDocument{
			DocID: docID,
			Score: score,
		})
	}
	sort.Slice(sortedScoredDocs, func(i, j int) bool {
		return less(sortedScoredDocs[i], sortedScoredDocs[j])
	})

	return sortedScoredDocs, nil
}
