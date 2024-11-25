package engine

// # TODOs
//
// - Implement new queries.
// - Modularize query execution.

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

type queryEngine struct {
	segments  []*storage.Segment
	totalDocs uint32
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
	block    *storage.Block
	iterator storage.PostingListIterator
	docID    uint32
}

// minBlockHeap implements heap.Interface for processing blocks in sorted order by docID
type minBlockHeap []*blockEntry

func (h minBlockHeap) Len() int           { return len(h) }
func (h minBlockHeap) Less(i, j int) bool { return h[i].docID < h[j].docID }
func (h minBlockHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minBlockHeap) Push(x interface{}) {
	*h = append(*h, x.(*blockEntry))
}

func (h *minBlockHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// termBlockHeap manages blocks for a single term's iterator
type termBlockHeap struct {
	term   string
	blocks *minBlockHeap
}

func (th *termBlockHeap) Init() {
	th.blocks = &minBlockHeap{}
	heap.Init(th.blocks)
}

// MultiTermQuery executes a query for multiple terms and returns scored documents
func (qe *queryEngine) MultiTermQuery(terms []string, less func(doc1, doc2 ScoredDocument) bool) ([]ScoredDocument, error) {
	// Initialize heaps for each term
	termBlockHeaps, err := initializeTermHeaps(terms, qe.segments)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize term heaps: %w", err)
	}

	var scoredDocuments []ScoredDocument

	for {
		matchingDocID, found := findMatchingDocument(termBlockHeaps)
		if !found {
			break
		}

		score, err := qe.calculateScore(termBlockHeaps, matchingDocID)
		if err != nil {
			return nil, fmt.Errorf("error calculating score: %w", err)
		}

		scoredDocuments = append(scoredDocuments, ScoredDocument{
			DocID: matchingDocID,
			Score: score,
		})

		// Advance all heaps after scoring
		for _, termBlockHeap := range termBlockHeaps {
			if err := advanceTermHeap(termBlockHeap); err != nil {
				return nil, fmt.Errorf("error advancing heap: %w", err)
			}
		}
	}

	// Sort results
	sort.Slice(scoredDocuments, func(i, j int) bool {
		return less(scoredDocuments[i], scoredDocuments[j])
	})

	return scoredDocuments, nil
}

// initializeTermHeaps creates and initializes heaps for each term
func initializeTermHeaps(terms []string, segments []*storage.Segment) ([]*termBlockHeap, error) {
	var termHeaps []*termBlockHeap

	for _, term := range terms {
		heap := &termBlockHeap{
			term: term,
		}
		heap.Init()

		// Get iterators for this term from all segments
		for _, segment := range segments {
			iterator, err := segment.TermIterator(term)
			if err != nil {
				return nil, fmt.Errorf("error creating iterator for term %s: %w", term, err)
			}

			if _, ok := iterator.(*storage.EmptyIterator); ok {
				continue
			}

			// Add initial block
			hasNext, err := iterator.Next()
			if err != nil {
				return nil, fmt.Errorf("error advancing iterator for term %s: %w", term, err)
			}
			if hasNext {
				docID, err := iterator.DocID()
				if err != nil {
					return nil, fmt.Errorf("error getting docID for term %s: %w", term, err)
				}
				block := iterator.CurrentBlock()
				heap.blocks.Push(&blockEntry{
					block:    block,
					iterator: iterator,
					docID:    docID,
				})
			}
		}

		if heap.blocks.Len() > 0 {
			termHeaps = append(termHeaps, heap)
		}
	}

	return termHeaps, nil
}

// findMatchingDocument finds the next document that contains all query terms
func findMatchingDocument(termHeaps []*termBlockHeap) (uint32, bool) {
	for {
		// Get smallest docID from heap tops
		smallestDocID := uint32(math.MaxUint32)
		hasMore := false

		for _, heap := range termHeaps {
			if heap.blocks.Len() > 0 {
				topDocID := (*heap.blocks)[0].docID
				if topDocID < smallestDocID {
					smallestDocID = topDocID
					hasMore = true
				}
			}
		}

		if !hasMore {
			return 0, false // No more documents
		}

		// Check if all heaps have this docID at top
		allMatch := true
		for _, heap := range termHeaps {
			if heap.blocks.Len() == 0 || (*heap.blocks)[0].docID != smallestDocID {
				allMatch = false
				break
			}
		}

		if allMatch {
			return smallestDocID, true
		}

		// Advance heaps that have the smallest docID
		for _, heap := range termHeaps {
			if heap.blocks.Len() > 0 && (*heap.blocks)[0].docID == smallestDocID {
				if err := advanceTermHeap(heap); err != nil {
					continue
				}
			}
		}
	}
}

// advanceTermHeap advances the top entry in the term's heap
func advanceTermHeap(th *termBlockHeap) error {
	if th.blocks.Len() == 0 {
		return nil
	}

	entry := heap.Pop(th.blocks).(*blockEntry)
	hasNext, err := entry.iterator.Next()
	if err != nil {
		return fmt.Errorf("error advancing iterator: %w", err)
	}

	if hasNext {
		docID, err := entry.iterator.DocID()
		if err != nil {
			return fmt.Errorf("error getting next docID: %w", err)
		}
		entry.docID = docID
		heap.Push(th.blocks, entry)
	}

	return nil
}

// calculateScore computes the TF-IDF score for a document
func (qe *queryEngine) calculateScore(termHeaps []*termBlockHeap, docID uint32) (float64, error) {
	var score float64

	for _, th := range termHeaps {
		if th.blocks.Len() == 0 {
			continue
		}

		entry := (*th.blocks)[0]
		if entry.docID != docID {
			continue
		}

		termFrequency, err := entry.iterator.TermFrequency()
		if err != nil {
			return 0, fmt.Errorf("error getting term frequency: %w", err)
		}

		// Calculate document frequency
		documentFrequency := 0
		for _, segment := range qe.segments {
			if metadata, exists := segment.Terms[th.term]; exists {
				documentFrequency += int(metadata.TotalDocs)
			}
		}

		// Calculate TF-IDF score component
		idf := math.Log(float64(qe.totalDocs+1) / float64(documentFrequency+1))
		score += float64(termFrequency) * idf
	}

	return score, nil
}
