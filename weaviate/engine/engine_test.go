package engine

import (
	"math"
	"testing"
	"weaviate/fetcher"
	"weaviate/storage"
)

// createMockSegment initializes a mock segment using a slice of TermPosting structs.
func createMockSegment(postings []fetcher.TermPosting) *storage.Segment {
	segment := storage.NewSegment()
	_ = segment.BulkIndex(postings)
	return segment
}

// countUniqueDocs computes the number of unique documents in the provided TermPosting slice.
func countUniqueDocs(postings []fetcher.TermPosting) uint32 {
	docIDSet := make(map[uint32]struct{})
	for _, posting := range postings {
		docIDSet[posting.DocID] = struct{}{}
	}
	return uint32(len(docIDSet))
}

func TestSingleTermQuery(t *testing.T) {
	postings := []fetcher.TermPosting{
		{Term: "anakin", DocID: 1, TermFrequency: 1.0},
		{Term: "anakin", DocID: 2, TermFrequency: 2.0},
		{Term: "anakin", DocID: 3, TermFrequency: 0.5},
	}

	segment := createMockSegment(postings)
	totalDocs := countUniqueDocs(postings)

	queryEngine, err := NewQueryEngine([]*storage.Segment{segment}, totalDocs)
	if err != nil {
		t.Fatalf("Failed to initialize QueryEngine: %v", err)
	}

	results, err := queryEngine.MultiTermQuery([]string{"anakin"}, func(d1, d2 ScoredDocument) bool {
		return d1.Score > d2.Score
	})
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	expectedDocCount := 3
	if len(results) != expectedDocCount {
		t.Fatalf("Expected %d results, got %d", expectedDocCount, len(results))
	}

	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Fatalf("Results are not sorted by score")
		}
	}
}

func TestMultiTermQuery(t *testing.T) {
	tfJedi := float32(2.0)
	tfSith := float32(3.0)
	postings := []fetcher.TermPosting{
		{Term: "jedi", DocID: 1, TermFrequency: 1.0},
		{Term: "jedi", DocID: 2, TermFrequency: tfJedi},
		{Term: "sith", DocID: 2, TermFrequency: tfSith},
		{Term: "sith", DocID: 3, TermFrequency: 0.5},
	}

	segment := createMockSegment(postings)
	totalDocs := countUniqueDocs(postings)

	queryEngine, err := NewQueryEngine([]*storage.Segment{segment}, totalDocs)
	if err != nil {
		t.Fatalf("Failed to initialize QueryEngine: %v", err)
	}

	results, err := queryEngine.MultiTermQuery([]string{"jedi", "sith"}, func(d1, d2 ScoredDocument) bool {
		return d1.Score > d2.Score
	})
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// Only DocID 2 matches both terms
	expectedDocCount := 1
	if len(results) != expectedDocCount {
		t.Fatalf("Expected %d results, got %d", expectedDocCount, len(results))
	}

	expectedDocID := uint32(2)
	if results[0].DocID != expectedDocID {
		t.Errorf("Expected DocID %d, got %d", expectedDocID, results[0].DocID)
	}

	dfJedi := 2
	dfSith := 2
	idfJedi := math.Log(float64(totalDocs+1) / float64(dfJedi+1))
	idfSith := math.Log(float64(totalDocs+1) / float64(dfSith+1))
	expectedScore := float64(tfJedi)*float64(idfJedi) + float64(tfSith)*float64(idfSith)

	if math.Abs(results[0].Score-expectedScore) > 1e-2 {
		t.Errorf("Expected score %.2f, got %.2f", expectedScore, results[0].Score)
	}
}

// TestEmptyQueryEngine tests QueryEngine initialization with no segments.
func TestEmptyQueryEngine(t *testing.T) {
	_, err := NewQueryEngine([]*storage.Segment{}, 100)
	if err == nil {
		t.Fatalf("Expected error when initializing QueryEngine with no segments")
	}
}

// TestMultiSegmentQuery tests the QueryEngine with multiple segments.
func TestMultiSegmentQuery(t *testing.T) {
	postings1 := []fetcher.TermPosting{
		{Term: "rebels", DocID: 1, TermFrequency: 1.0},
		{Term: "rebels", DocID: 2, TermFrequency: 2.0},
	}
	postings2 := []fetcher.TermPosting{
		{Term: "empire", DocID: 3, TermFrequency: 3.0},
		{Term: "empire", DocID: 4, TermFrequency: 0.5},
	}

	segment1 := createMockSegment(postings1)
	segment2 := createMockSegment(postings2)

	totalDocs := countUniqueDocs(append(postings1, postings2...))

	queryEngine, err := NewQueryEngine([]*storage.Segment{segment1, segment2}, totalDocs)
	if err != nil {
		t.Fatalf("Failed to initialize QueryEngine: %v", err)
	}

	results, err := queryEngine.MultiTermQuery([]string{"rebels", "empire"}, func(d1, d2 ScoredDocument) bool {
		return d1.Score > d2.Score
	})
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// There should be no results because no document contains both terms across segments
	if len(results) != 0 {
		t.Fatalf("Expected 0 results, got %d", len(results))
	}
}

func TestMultiSegmentQueryWithManyTerms(t *testing.T) {
	postings1 := []fetcher.TermPosting{
		{Term: "rebels", DocID: 1, TermFrequency: 1.0}, // match "rebels" DocID: 1
		{Term: "rebels", DocID: 2, TermFrequency: 2.0},
		{Term: "alliance", DocID: 1, TermFrequency: 1.5},
		{Term: "empire", DocID: 1, TermFrequency: 1.0}, // match "empire" DocID: 1
		{Term: "hope", DocID: 1, TermFrequency: 1.5},   // match "hope" DocID: 1
		{Term: "jedi", DocID: 2, TermFrequency: 2.0},
		{Term: "hope", DocID: 2, TermFrequency: 1.0},
		{Term: "darkside", DocID: 2, TermFrequency: 1.0},
	}

	postings2 := []fetcher.TermPosting{
		{Term: "rebels", DocID: 6, TermFrequency: 2.5}, // match "rebels" DocID: 6
		{Term: "empire", DocID: 6, TermFrequency: 3.0}, // match "empire" DocID: 6
		{Term: "hope", DocID: 6, TermFrequency: 1.5},   // match "hope" DocID: 6
		{Term: "darkside", DocID: 6, TermFrequency: 1.5},
		{Term: "darkside", DocID: 7, TermFrequency: 0.5},
		{Term: "sith", DocID: 11, TermFrequency: 0.5},
	}

	segment1 := createMockSegment(postings1)
	segment2 := createMockSegment(postings2)
	totalDocs := countUniqueDocs(append(postings1, postings2...))

	queryEngine, err := NewQueryEngine([]*storage.Segment{segment1, segment2}, totalDocs)
	if err != nil {
		t.Fatalf("Failed to initialize QueryEngine: %v", err)
	}

	// Multi query term with all terms matching
	results, err := queryEngine.MultiTermQuery([]string{"rebels", "empire", "hope"}, func(d1, d2 ScoredDocument) bool {
		return d1.Score > d2.Score
	})
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	expectedDocIDs := []uint32{1, 6} // DocID 1 from Segment 1 and DocID 6 from Segment 2.
	if len(results) != len(expectedDocIDs) {
		t.Fatalf("Expected %d results, got %d", len(expectedDocIDs), len(results))
	}

	resultDocIDs := make(map[uint32]bool)
	for _, result := range results {
		resultDocIDs[result.DocID] = true
	}

	for _, expectedDocID := range expectedDocIDs {
		if !resultDocIDs[expectedDocID] {
			t.Errorf("Expected DocID %d in results, but it was missing", expectedDocID)
		}
	}

	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Fatalf("Results are not sorted by score: %+v", results)
		}
	}
}

func TestMultiSegmentQueryWithResults(t *testing.T) {
	postings1 := []fetcher.TermPosting{
		{Term: "rebels", DocID: 1, TermFrequency: 1.5},
		{Term: "empire", DocID: 1, TermFrequency: 2.0},
		{Term: "rebels", DocID: 2, TermFrequency: 2.0},
		{Term: "empire", DocID: 2, TermFrequency: 1.0},
	}

	postings2 := []fetcher.TermPosting{
		{Term: "rebels", DocID: 3, TermFrequency: 2.5},
		{Term: "empire", DocID: 3, TermFrequency: 1.5},
	}

	segment1 := createMockSegment(postings1)
	segment2 := createMockSegment(postings2)

	totalDocs := countUniqueDocs(append(postings1, postings2...))

	queryEngine, err := NewQueryEngine([]*storage.Segment{segment1, segment2}, totalDocs)
	if err != nil {
		t.Fatalf("Failed to initialize QueryEngine: %v", err)
	}

	results, err := queryEngine.MultiTermQuery([]string{"rebels", "empire"}, func(d1, d2 ScoredDocument) bool {
		return d1.Score > d2.Score
	})
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// DocIDs 1, 2, and 3 should match both terms.
	expectedDocIDs := []uint32{1, 2, 3}
	if len(results) != len(expectedDocIDs) {
		t.Fatalf("Expected %d results, got %d", len(expectedDocIDs), len(results))
	}

	resultDocIDs := make(map[uint32]bool)
	for _, result := range results {
		resultDocIDs[result.DocID] = true
	}

	for _, expectedDocID := range expectedDocIDs {
		if !resultDocIDs[expectedDocID] {
			t.Errorf("Expected DocID %d in results, but it was missing", expectedDocID)
		}
	}
}

// TestScoringFunction tests the QueryEngine's scoring function.
func TestScoringFunction(t *testing.T) {
	const (
		skywalkerRebelsTermFreq = 7.2  // "skywalker" term frequency in rebels document (DocID = 2)
		vaderEmpireTermFreq     = 12.5 // "vader" term frequency in rebels document (DocID = 2)
		expectedDocID           = 2    // The rebels document (DocID 2) mentions both "skywalker" and "vader"
	)

	postings := []fetcher.TermPosting{
		{Term: "skywalker", DocID: 1, TermFrequency: 2.0},
		{Term: "skywalker", DocID: expectedDocID, TermFrequency: skywalkerRebelsTermFreq},
		{Term: "vader", DocID: expectedDocID, TermFrequency: vaderEmpireTermFreq},
		{Term: "vader", DocID: 3, TermFrequency: 4.0},
	}

	segment := createMockSegment(postings)
	totalDocs := countUniqueDocs(postings)

	if segment.TotalDocs() != totalDocs {
		t.Fatalf("Invalid number of total documents, expected: %d got: %d\n", totalDocs, segment.TotalDocs())
	}

	queryEngine, err := NewQueryEngine([]*storage.Segment{segment}, totalDocs)
	if err != nil {
		t.Fatalf("Failed to initialize QueryEngine: %v", err)
	}

	results, err := queryEngine.MultiTermQuery([]string{"skywalker", "vader"}, func(d1, d2 ScoredDocument) bool {
		return d1.Score > d2.Score
	})
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	expectedScore := skywalkerRebelsTermFreq*math.Log(float64(totalDocs+1)/float64(2+1)) +
		vaderEmpireTermFreq*math.Log(float64(totalDocs+1)/float64(2+1))

	if results[0].DocID != uint32(expectedDocID) {
		t.Errorf("Expected DocID %d (Luke), got %d", uint32(expectedDocID), results[0].DocID)
	}

	if math.Abs(results[0].Score-expectedScore) > 1e-2 {
		t.Errorf("Expected score %.2f, got %.2f", expectedScore, results[0].Score)
	}
}
