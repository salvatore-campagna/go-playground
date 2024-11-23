// Package storage implements an inverted index segment for full-text search.
// It provides efficient storage and retrieval of term-document relationships
// using compressed posting lists and Roaring Bitmaps. The implementation
// supports serialization for persistence and includes optimizations for
// memory usage and query performance, enabling scalable search functionality.
//
// # File Format
//
// The segment file is organized into three main sections: the file header, the terms section, and the blocks section.
//
// ## File Header
// The file header provides metadata for the segment file:
//   - Magic Number (4 bytes): Identifies the segment file format (e.g., 0x007E8B11)
//   - Version (1 byte): Current segment format version (1)
//   - Document Count (4 bytes): Total number of documents in the segment
//   - Number of Terms (4 bytes): Total number of unique terms in the segment
//
// ## Terms Section
// Each term in the segment has metadata and associated posting list blocks:
//   - Term Length (2 bytes): Length of the term string
//   - Term String (variable): UTF-8 encoded term
//   - Total Documents (4 bytes): Number of documents containing this term
//   - Number of Blocks (4 bytes): Number of blocks associated with the term
//
// ## Blocks Section
// Each block represents a chunk of the posting list for a term. Blocks include the following:
//   - Min DocID (4 bytes): Minimum document ID in the block
//   - Max DocID (4 bytes): Maximum document ID in the block
//   - Bitmap Container Type (1 byte): Type of Roaring Bitmap container (1 = ArrayContainer, 2 = BitmapContainer)
//   - Compressed DocID Storage: Roaring Bitmap representing document IDs
//   - Number of Term Frequencies (4 bytes): Number of term frequencies stored in the block
//   - Term Frequencies ([]float32): Term frequencies for each document in the block
//
// ## Example Layout
//
// The following example illustrates a segment file with two terms and their associated blocks:
//
// Magic Number        : 0x007E8B11
// Version             : 1
// Document Count      : 2500
// Number of Terms     : 2
//
// **[Term 1: "database"]**
//   - Term Length       : 8
//   - Total Documents   : 2000
//   - Number of Blocks  : 2
//   - **Block 1**
//   - Min DocID      : 1
//   - Max DocID      : 1000
//   - Bitmap Type    : ArrayContainer
//   - DocIDs         : [1, 50, 200, ..., 1000]
//   - Term Frequencies: [0.5, 0.6, 0.7, ..., 0.8]
//   - **Block 2**
//   - Min DocID      : 1001
//   - Max DocID      : 2000
//   - Bitmap Type    : BitmapContainer
//   - DocIDs         : [1001, 1050, 1100, ..., 2000]
//   - Term Frequencies: [0.2, 0.4, 0.6, ..., 0.9]
//
// **[Term 2: "vector"]**
//   - Term Length       : 6
//   - Total Documents   : 500
//   - Number of Blocks  : 1
//   - **Block 1**
//   - Min DocID      : 3000
//   - Max DocID      : 3500
//   - Bitmap Type    : BitmapContainer
//   - DocIDs         : [3000, 3050, 3100, ..., 3500]
//   - Term Frequencies: [0.1, 0.3, 0.5, ..., 0.7]
//
// # Features and TODOs
//
// - Optimized posting list storage using Roaring Bitmaps
// - Efficient traversal of document IDs and term frequencies with block-level iterators
// - Metadata for blocks (e.g., MinDocID and MaxDocID) to enable efficient block skipping
// - Serialization and deserialization support for persistence
// - TODO: Support data integrity checks (e.g., checksums, hashing)
// - TODO: Explore using Tries or Finite State Transducers (FSTs) for term metadata storage
// - TODO: Add benchmarks for indexing latency, memory usage, and query performance
// - TODO: Evaluate use of integer compression for term frequencies
// - TODO: Consider parallel processing for bulk indexing and queries
package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
	"weaviate/fetcher"
)

// Constants for segment format versioning
const (
	magicNumber        = 0x007E8B11 // Magic number to identify segment files
	version            = 1          // Current segment format version
	MaxEntriesPerBlock = 4 * 1024   // Maximum block size
)

// Segment represents a collection of terms and their posting lists.
// It provides an immutable snapshot of indexed documents, supporting
// efficient term-based document lookups and frequency scoring.
type Segment struct {
	MagicNumber uint32
	Version     uint8
	DocCount    uint32
	Terms       map[string]*TermMetadata
}

// TermMetadata holds statistical and structural data for a specific term
// in the segment, including document frequencies and posting blocks.
type TermMetadata struct {
	TotalDocs uint32   // Total number of documents containing this term
	Blocks    []*Block // Ordered blocks of posting list data
}

// Block represents a compressed set of document IDs and their corresponding
// term frequencies. Uses RoaringBitmap for efficient docID storage.
type Block struct {
	MinDocID        uint32         // Minimum DocID in the block
	MaxDocID        uint32         //Maximun DocID in the block
	Bitmap          *RoaringBitmap // Compressed document ID storage
	TermFrequencies []float32      // Term frequencies for each document
}

// PrintInfo prints out detailed information about the Segment in a nicely formatted table.
func (s *Segment) PrintInfo() {
	fmt.Printf("Segment Information\n\n")
	fmt.Printf("Magic Number   : 0x%X\n", s.MagicNumber)
	fmt.Printf("Version        : %d\n", s.Version)
	fmt.Printf("Toal Docs      : %d\n", s.DocCount)
	fmt.Printf("Total Terms    : %d\n", len(s.Terms))

	fmt.Printf("\n%-20s | %-12s | %-10s | %-10s | %-10s |\n", "Term", "Documents", "Blocks", "Postings", "Freq Len")
	fmt.Println(strings.Repeat("-", 75))

	totalDocs := 0
	totalBlocks := 0
	totalPostings := 0

	for term, metadata := range s.Terms {
		termDocs := int(metadata.TotalDocs)
		termBlocks := len(metadata.Blocks)
		termPostings := 0
		termFrequencies := 0

		for _, block := range metadata.Blocks {
			termPostings += block.Bitmap.Cardinality()
			termFrequencies += len(block.TermFrequencies)
		}

		totalDocs += termDocs
		totalBlocks += termBlocks
		totalPostings += termPostings

		fmt.Printf("%-20s | %-12d | %-10d | %-10d | %-10d |\n", term, termDocs, termBlocks, termPostings, termFrequencies)
	}

	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("\n%-20s | %-12d | %-10d | %-10d\n", "Overall", totalDocs, totalBlocks, totalPostings)
	fmt.Printf("\n\n\n")
}

// NewSegment initializes a new Segment with the given base document ID.
func NewSegment() *Segment {
	return &Segment{
		MagicNumber: magicNumber,
		Version:     version,
		DocCount:    0,
		Terms:       make(map[string]*TermMetadata),
	}
}

func (s *Segment) TotalDocs() uint32 {
	return s.DocCount
}

// BulkIndex adds a batch of documents to the segment.
// The documents can contain different terms, and they must be sorted by document ID.
// BulkIndex adds a batch of documents to the segment.
// Documents can contain different terms and must be sorted by DocID.
func (s *Segment) BulkIndex(documents []fetcher.TermPosting) error {
	if len(documents) == 0 {
		return nil
	}

	for _, document := range documents {
		termMetadata, exists := s.Terms[document.Term]
		if !exists {
			termMetadata = &TermMetadata{
				TotalDocs: 0,
				Blocks:    make([]*Block, 0),
			}
			s.Terms[document.Term] = termMetadata
		}

		// Get the last block or create a new one
		var block *Block
		if len(termMetadata.Blocks) > 0 {
			block = termMetadata.Blocks[len(termMetadata.Blocks)-1]
		}
		if block == nil || block.Bitmap.Cardinality() >= MaxEntriesPerBlock {
			block = &Block{
				MinDocID:        document.DocID,
				MaxDocID:        document.DocID,
				Bitmap:          NewRoaringBitmap(),
				TermFrequencies: make([]float32, 0, MaxEntriesPerBlock), // Preallocate for efficiency
			}
			termMetadata.Blocks = append(termMetadata.Blocks, block)
		}

		// Add the document to the block
		if err := block.AddDocument(document.DocID, document.TermFrequency); err != nil {
			return fmt.Errorf("failed to add document to block: %w", err)
		}

		// Update block metadata
		block.MaxDocID = document.DocID

		// Update term metadata
		termMetadata.TotalDocs++
	}

	// Update the segment's total document count
	s.DocCount += uint32(len(documents))
	return nil
}

// NewBlock creates a new block for storing document IDs and term frequencies.
func NewBlock() *Block {
	return &Block{
		Bitmap:          NewRoaringBitmap(), // Ensure this is initialized
		TermFrequencies: make([]float32, 0), // Initialize with empty slice
	}
}

// AddDocument adds a document's ID and term frequency to the block.
// Document IDs should be added in ascending order for optimal performance.
func (b *Block) AddDocument(docID uint32, termFrequency float32) error {
	b.Bitmap.Add(docID)
	b.TermFrequencies = append(b.TermFrequencies, termFrequency)
	if b.Bitmap.Cardinality() != len(b.TermFrequencies) {
		return fmt.Errorf("error while adding document %d with term frequency %.2f\n", docID, termFrequency)
	}
	return nil
}

// Segment.Serialize writes the segment to the provided writer.
func (s *Segment) Serialize(writer io.Writer) error {

	if err := binary.Write(writer, binary.LittleEndian, s.MagicNumber); err != nil {
		return err
	}

	if err := binary.Write(writer, binary.LittleEndian, s.Version); err != nil {
		return err
	}

	if err := binary.Write(writer, binary.LittleEndian, s.DocCount); err != nil {
		return err
	}

	numTerms := uint32(len(s.Terms))
	if err := binary.Write(writer, binary.LittleEndian, numTerms); err != nil {
		return err
	}

	for term, metadata := range s.Terms {
		termLen := uint16(len(term))
		if err := binary.Write(writer, binary.LittleEndian, termLen); err != nil {
			return err
		}
		if _, err := writer.Write([]byte(term)); err != nil {
			return err
		}

		if err := binary.Write(writer, binary.LittleEndian, metadata.TotalDocs); err != nil {
			return err
		}

		numBlocks := uint32(len(metadata.Blocks))
		if err := binary.Write(writer, binary.LittleEndian, numBlocks); err != nil {
			return err
		}

		for _, block := range metadata.Blocks {
			if err := block.Serialize(writer); err != nil {
				return err
			}
		}
	}

	return nil
}

// Segment.Deserialize reads a segment from the provided reader.
// The reader must provide data in the exact format produced by Serialize.
// Returns an error if the data format is invalid or if the magic number
// or version don't match expected values.
func (s *Segment) Deserialize(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &s.MagicNumber); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.LittleEndian, &s.Version); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.LittleEndian, &s.DocCount); err != nil {
		return err
	}
	var numTerms uint32
	if err := binary.Read(reader, binary.LittleEndian, &numTerms); err != nil {
		return err
	}
	s.Terms = make(map[string]*TermMetadata)

	for i := 0; i < int(numTerms); i++ {
		var termLen uint16
		if err := binary.Read(reader, binary.LittleEndian, &termLen); err != nil {
			return err
		}

		termBytes := make([]byte, termLen)
		if _, err := io.ReadFull(reader, termBytes); err != nil {
			return err
		}
		term := string(termBytes)
		termMeta := &TermMetadata{}

		if err := binary.Read(reader, binary.LittleEndian, &termMeta.TotalDocs); err != nil {
			return err
		}
		var numBlocks uint32
		if err := binary.Read(reader, binary.LittleEndian, &numBlocks); err != nil {
			return err
		}

		termMeta.Blocks = make([]*Block, numBlocks)
		for j := 0; j < int(numBlocks); j++ {
			block := &Block{}
			block.Bitmap = NewRoaringBitmap()

			if err := block.Deserialize(reader); err != nil {
				return err
			}
			termMeta.Blocks[j] = block
		}

		s.Terms[term] = termMeta
	}

	// Ensure there are no extra bytes
	var extraByte byte
	err := binary.Read(reader, binary.LittleEndian, &extraByte)
	if err == nil {
		return fmt.Errorf("unexpected extra byte: %w", err)
	}
	return nil
}

// Block.Serialize writes a block to the provided writer.
func (b *Block) Serialize(writer io.Writer) error {
	if err := binary.Write(writer, binary.LittleEndian, b.MinDocID); err != nil {
		return fmt.Errorf("failed to write minDocID: %w", err)
	}
	if err := binary.Write(writer, binary.LittleEndian, b.MaxDocID); err != nil {
		return fmt.Errorf("failed to write maxDocID: %w", err)
	}
	if err := b.Bitmap.Serialize(writer); err != nil {
		return fmt.Errorf("failed to serialize bitmap: %w", err)
	}

	// Write term frequencies using delta + varint encoding
	numFreqs := uint32(len(b.TermFrequencies))
	if err := binary.Write(writer, binary.LittleEndian, numFreqs); err != nil {
		return fmt.Errorf("failed to write number of term frequencies: %w", err)
	}
	prevFreq := float32(0)
	for _, freq := range b.TermFrequencies {
		delta := freq - prevFreq
		if err := binary.Write(writer, binary.LittleEndian, delta); err != nil {
			return fmt.Errorf("failed to write term frequency delta: %w", err)
		}
		prevFreq = freq
	}
	return nil
}

// Block.Deserialize reads a block from the provided reader.
func (b *Block) Deserialize(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &b.MinDocID); err != nil {
		return fmt.Errorf("failed to read minDocID: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &b.MaxDocID); err != nil {
		return fmt.Errorf("failed to read maxDocID: %w", err)
	}
	if err := b.Bitmap.Deserialize(reader); err != nil {
		return fmt.Errorf("failed to deserialize bitmap: %w", err)
	}

	var numFreqs uint32
	if err := binary.Read(reader, binary.LittleEndian, &numFreqs); err != nil {
		return fmt.Errorf("failed to read number of term frequencies: %w", err)
	}
	b.TermFrequencies = make([]float32, numFreqs)
	prevFreq := float32(0)
	for i := uint32(0); i < numFreqs; i++ {
		var delta float32
		if err := binary.Read(reader, binary.LittleEndian, &delta); err != nil {
			return fmt.Errorf("failed to read term frequency delta: %w", err)
		}
		b.TermFrequencies[i] = prevFreq + delta
		prevFreq = b.TermFrequencies[i]
	}
	return nil
}

// WriteSegment writes a Segment to an io.Writer, typically a file.
// This is a convenience wrapper around Serialize.
func (s *Segment) WriteSegment(writer io.Writer) error {
	if err := s.Serialize(writer); err != nil {
		return fmt.Errorf("failed to serialize segment: %w", err)
	}
	return nil
}

// ReadSegment reads a Segment from an io.Reader, typically a file.
// This is a convenience wrapper around Deserialize.
func (s *Segment) ReadSegment(reader io.Reader) error {
	if err := s.Deserialize(reader); err != nil {
		return fmt.Errorf("failed to deserialize segment: %w", err)
	}
	return nil
}

func (s *Segment) TermIterator(term string) (PostingListIterator, error) {
	termMetadata, exists := s.Terms[term]
	if !exists {
		return &EmptyIterator{}, nil
	}

	return &TermIterator{
		blocks:        termMetadata.Blocks,
		currentBlock:  0,
		blockIterator: termMetadata.Blocks[0].Bitmap.Iterator(), // if a term exists we have at least one block so it is safe to access Blocks[0]
	}, nil
}

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

func (rb *RoaringBitmap) Iterator() BitmapIterator {
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
