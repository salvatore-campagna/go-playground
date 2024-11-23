// Package engine provides the implementation for a query engine that processes
// multi-term queries across multiple index segments. It supports scoring documents
// using TF-IDF and efficiently retrieves ranked documents using a priority queue (min-heap).
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
	DocID uint32  // The unique identifier for the document.
	Score float64 // The computed relevance score for the document.
}

// QueryEngine defines the interface for executing queries on an index.
type QueryEngine interface {
	// MultiTermQuery performs a query on multiple terms and ranks the results using the provided comparator.
	// terms: List of terms to search for.
	// less: A comparator function to sort the scored documents.
	// Returns a slice of ScoredDocuments or an error if the query fails.
	MultiTermQuery(terms []string, less func(i, j ScoredDocument) bool) ([]ScoredDocument, error)
}

// queryEngine is the default implementation of the QueryEngine interface.
type queryEngine struct {
	segments  []*storage.Segment // List of index segments to query.
	totalDocs uint32             // Total number of documents across all segments.
}

// NewQueryEngine initializes a QueryEngine with the given segments and total document count.
// Returns an error if the input is invalid.
func NewQueryEngine(segments []*storage.Segment, totalDocs uint32) (QueryEngine, error) {
	if segments == nil || len(segments) <= 0 {
		return nil, fmt.Errorf("no segment to query")
	}

	if totalDocs <= 0 {
		return nil, fmt.Errorf("invalid number of documents: %d", totalDocs)
	}

	return &queryEngine{
		segments:  segments,
		totalDocs: totalDocs,
	}, nil
}

// heapEntry represents an entry in the min-heap used during query processing.
type heapEntry struct {
	term     string                      // The term associated with this entry.
	iterator storage.PostingListIterator // Iterator over the posting list for the term.
	docID    uint32                      // The current document ID from the iterator.
}

// minHeap is a priority queue for heapEntries, sorted by docID.
type minHeap []*heapEntry

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool { return h[i].docID < h[j].docID }

func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(value interface{}) {
	*h = append(*h, value.(*heapEntry))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[0 : n-1]
	return item
}

// getTermPostingListIterators retrieves posting list iterators for the given terms from all segments.
// Returns a map of terms to their iterators or an error if any term lookup fails.
func getTermPostingListIterators(terms []string, segments []*storage.Segment) (map[string][]storage.PostingListIterator, error) {
	termIterators := make(map[string][]storage.PostingListIterator)
	for _, term := range terms {
		var iterators []storage.PostingListIterator
		for _, segment := range segments {
			iter, err := segment.TermIterator(term)
			if err != nil {
				return nil, err
			}
			iterators = append(iterators, iter)
		}
		termIterators[term] = iterators
	}
	return termIterators, nil
}

// MultiTermQuery processes a multi-term query across all segments.
// Computes the TF-IDF scores for documents that match all query terms and ranks them using the provided comparator.
func (qe *queryEngine) MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error) {
	termIterators, err := getTermPostingListIterators(terms, qe.segments)
	if err != nil {
		return nil, err
	}

	minHeap := &minHeap{}
	heap.Init(minHeap)

	// Initialize the heap with the first document for each term.
	for term, iters := range termIterators {
		for _, iter := range iters {
			hasNext, err := iter.Next()
			if err != nil {
				return nil, err
			}
			if hasNext {
				docID, err := iter.DocID()
				if err != nil {
					return nil, err
				}
				heap.Push(minHeap, &heapEntry{term, iter, docID})
			}
		}
	}

	scoredDocs := make(map[uint32]float64)
	docTermCounts := make(map[uint32]int)
	termDocTFs := make(map[string]map[uint32]float32)

	// Initialize term-document frequency maps.
	for _, term := range terms {
		termDocTFs[term] = make(map[uint32]float32)
	}

	// Process the heap until it is empty.
	for minHeap.Len() > 0 {
		entry := heap.Pop(minHeap).(*heapEntry)
		docID := entry.docID
		docTermCounts[docID]++

		tf, err := entry.iterator.TermFrequency()
		if err != nil {
			return nil, err
		}
		termDocTFs[entry.term][docID] = tf

		// Score the document if it matches all terms.
		if docTermCounts[docID] == len(terms) {
			totalScore := 0.0
			for _, termFrequencyMap := range termDocTFs {
				tf := termFrequencyMap[docID]
				df := len(termFrequencyMap)
				idf := math.Log(float64(qe.totalDocs) / float64(df+1))
				tfidf := float64(tf) * idf
				totalScore += tfidf
			}
			scoredDocs[docID] = totalScore
		}

		// Push the next document for this term into the heap.
		hasNext, err := entry.iterator.Next()
		if err != nil {
			return nil, err
		}
		if hasNext {
			docID, err := entry.iterator.DocID()
			if err != nil {
				return nil, err
			}
			heap.Push(minHeap, &heapEntry{entry.term, entry.iterator, docID})
		}
	}

	var results []ScoredDocument
	for docID, score := range scoredDocs {
		results = append(results, ScoredDocument{
			DocID: docID,
			Score: score,
		})
	}

	// Sort documents
	sort.Slice(results, func(i, j int) bool {
		return less(results[i], results[j])
	})
	return results, nil
}
