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
	MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error)
}

// queryEngine is the default implementation of the QueryEngine interface.
type queryEngine struct {
	segments  []*storage.Segment // List of index segments to query.
	totalDocs uint32             // Total number of documents across all segments.
}

// NewQueryEngine initializes a QueryEngine with the given segments and total document count.
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

// minBlockHeap is a priority queue (min-heap) of block entries.
type minBlockHeap []*blockEntry

func (h minBlockHeap) Len() int { return len(h) }

func (h minBlockHeap) Less(i, j int) bool {
	return h[i].block.MinDocID < h[j].block.MinDocID
}

func (h minBlockHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *minBlockHeap) Push(x interface{}) {
	*h = append(*h, x.(*blockEntry))
}

func (h *minBlockHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// getTermPostingListIterators retrieves posting list iterators for the given terms from all segments.
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

// MultiTermQuery processes a multi-term query across all segments.
func (qe *queryEngine) MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error) {
	termIterators, err := getTermPostingListIterators(terms, qe.segments)
	if err != nil {
		return nil, err
	}

	// Initialize the heap for blocks
	blockHeap := &minBlockHeap{}
	heap.Init(blockHeap)

	// Push initial blocks into the heap for each term
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

	scoredDocs := make(map[uint32]float64)
	docTermCounts := make(map[uint32]int)
	termDocTFs := make(map[string]map[uint32]float32)

	// Initialize term-document frequency maps
	for _, term := range terms {
		termDocTFs[term] = make(map[uint32]float32)
	}

	// Process the heap
	for blockHeap.Len() > 0 {
		entry := heap.Pop(blockHeap).(*blockEntry)
		docID := entry.docID
		docTermCounts[docID]++

		tf, err := entry.iterator.TermFrequency()
		if err != nil {
			return nil, fmt.Errorf("error retrieving TermFrequency: %v", err)
		}
		termDocTFs[entry.iterator.Term()][docID] = tf

		// Score the document if it matches all terms
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

		// Push the next document for this term into the heap
		hasNext, err := entry.iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error advancing iterator: %v", err)
		}
		if hasNext {
			docID, err := entry.iterator.DocID()
			if err != nil {
				return nil, fmt.Errorf("error retrieving next DocID: %v", err)
			}
			heap.Push(blockHeap, &blockEntry{
				block:    entry.block,
				iterator: entry.iterator,
				docID:    docID,
			})
		}
	}

	// TODO: maybe use anothe rheap here too?
	var results []ScoredDocument
	for docID, score := range scoredDocs {
		results = append(results, ScoredDocument{
			DocID: docID,
			Score: score,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return less(results[i], results[j])
	})

	return results, nil
}
