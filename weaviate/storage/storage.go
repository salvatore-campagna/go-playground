package storage

// # TODOs
//
// - Add support for data integrity checks (e.g., checksums, hashing).
// - Explore using Tries or Finite State Transducers (FSTs) for term metadata storage to improve lookup efficiency.
// - Add benchmarks for indexing latency, memory usage, and query performance.
// - Evaluate the use of integer compression for term frequencies to reduce storage space.
// - Support dynamic updates to segments, including deletions and incremental additions.
// - Improve block skipping strategies for large posting lists to enhance query speed.
// - Explore using SIMD (Single Instruction, Multiple Data) techniques for accelerating operations on posting lists.
// - Extend support for storing additional metadata, such as document scores.

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"weaviate/fetcher"
)

// Constants for segment format versioning
const (
	magicNumber          = 0x007E8B11
	version              = 1
	MaxDcoumentsPerBlock = 16 * 1024
)

// Segment represents a collection of terms and their posting lists.
// It provides an immutable snapshot of indexed documents, supporting
// efficient term-based document lookups and frequency scoring.
type Segment struct {
	MagicNumber uint32
	Version     uint8
	DocIDs      *RoaringBitmap
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
	MaxDocID        uint32         // Maximun DocID in the block (not used, consider removing)
	Bitmap          *RoaringBitmap // Compressed document ID storage
	TermFrequencies []float32      // Term frequencies for each document
}

// PrintInfo prints out detailed information about the Segment.
func (s *Segment) PrintInfo() {
	fmt.Printf("Segment Information\n\n")
	fmt.Printf("Magic Number   : 0x%X\n", s.MagicNumber)
	fmt.Printf("Version        : %d\n", s.Version)
	fmt.Printf("Total Docs     : %d\n", s.DocIDs.Cardinality())
	fmt.Printf("Total Terms    : %d\n", len(s.Terms))

	// Term-level summary
	fmt.Printf("\n%-25s | %-15s | %-12s | %-12s |\n", "Term", "Documents", "Blocks", "Postings")
	fmt.Println(strings.Repeat("-", 70))

	totalDocs := 0
	totalBlocks := 0
	totalPostings := 0

	for term, metadata := range s.Terms {
		termDocs := int(metadata.TotalDocs)
		termBlocks := len(metadata.Blocks)
		termPostings := 0

		for _, block := range metadata.Blocks {
			termPostings += block.Bitmap.Cardinality()
		}

		totalDocs += termDocs
		totalBlocks += termBlocks
		totalPostings += termPostings

		fmt.Printf("%-25s | %-15d | %-12d | %-12d |\n", term, termDocs, termBlocks, termPostings)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("\n%-25s | %-15d | %-12d | %-12d\n", "Overall", totalDocs, totalBlocks, totalPostings)

	// Block-level summary
	fmt.Printf("\nDetailed Block Summary\n")
	fmt.Printf("%-25s | %-8s | %-10s | %-10s | %-12s | %-12s |\n", "Term", "Block", "MinDocID", "MaxDocID", "Cardinality", "FreqLen")
	fmt.Println(strings.Repeat("-", 94))

	for term, metadata := range s.Terms {
		termCardinality := 0
		termFreqLen := 0

		for i, block := range metadata.Blocks {
			blockCardinality := block.Bitmap.Cardinality()
			freqLen := len(block.TermFrequencies)

			termCardinality += blockCardinality
			termFreqLen += freqLen

			fmt.Printf("%-25s | %-8d | %-10d | %-10d | %-12d | %-12d |\n",
				term, i+1, block.MinDocID, block.MaxDocID, blockCardinality, freqLen)
		}

		fmt.Printf("%-25s | %-8s | %-10s | %-10s | %-12d | %-12d |\n", term, "Total", "-", "-", termCardinality, termFreqLen)
		fmt.Println(strings.Repeat("-", 94))
	}
}

// NewSegment initializes a new Segment with the given base document ID.
func NewSegment() *Segment {
	return &Segment{
		MagicNumber: magicNumber,
		Version:     version,
		DocIDs:      NewRoaringBitmap(),
		Terms:       make(map[string]*TermMetadata),
	}
}

// TotalDocs returns the total number of documents in the segment.
func (s *Segment) TotalDocs() uint32 {
	return uint32(s.DocIDs.Cardinality())
}

// BulkIndex adds a batch of terms to the segment.
func (s *Segment) BulkIndex(termPostings []fetcher.TermPosting) error {
	if len(termPostings) == 0 {
		return nil
	}

	for _, termPosting := range termPostings {
		if !s.DocIDs.Contains(termPosting.DocID) {
			s.DocIDs.Add(termPosting.DocID)
		}

		termMetadata, exists := s.Terms[termPosting.Term]
		if !exists {
			termMetadata = &TermMetadata{
				TotalDocs: 0,
				Blocks:    make([]*Block, 0),
			}
			s.Terms[termPosting.Term] = termMetadata
		}

		// Get the last block or create a new one if the current block is full
		var block *Block
		if len(termMetadata.Blocks) > 0 {
			block = termMetadata.Blocks[len(termMetadata.Blocks)-1]
		}
		if block == nil || block.Bitmap.Cardinality() >= MaxDcoumentsPerBlock {
			block = &Block{
				MinDocID:        termPosting.DocID,
				MaxDocID:        termPosting.DocID,
				Bitmap:          NewRoaringBitmap(),
				TermFrequencies: make([]float32, 0),
			}
			termMetadata.Blocks = append(termMetadata.Blocks, block)
		}

		// Add the document to the block (if does not already exist)
		if !block.Bitmap.Contains(termPosting.DocID) {
			if err := block.AddDocument(termPosting.DocID, termPosting.TermFrequency); err != nil {
				return fmt.Errorf("failed to add document to block: %w", err)
			}
			termMetadata.TotalDocs++
		}
	}

	return nil
}

// NewBlock creates a new block for storing document IDs and term frequencies.
func NewBlock() *Block {
	return &Block{
		Bitmap:          NewRoaringBitmap(),
		TermFrequencies: make([]float32, 0),
	}
}

// AddDocument adds a document's ID and term frequency to the block.
func (b *Block) AddDocument(docID uint32, termFrequency float32) error {
	b.Bitmap.Add(docID)
	// Keep track of the MinDocID and MaxDocID
	if b.Bitmap.Cardinality() == 1 {
		b.MinDocID = docID
		b.MaxDocID = docID
	} else {
		if docID < b.MinDocID {
			b.MinDocID = docID
		}
		if docID > b.MaxDocID {
			b.MaxDocID = docID
		}
	}

	// Add term frequency
	b.TermFrequencies = append(b.TermFrequencies, termFrequency)

	// Sanity check for consistency
	if b.Bitmap.Cardinality() != len(b.TermFrequencies) {
		return fmt.Errorf("mismatch between bitmap cardinality and term frequencies")
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
	if err := s.DocIDs.Serialize(writer); err != nil {
		return fmt.Errorf("failed to serialize DocIDs bitmap: %w", err)
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
func (s *Segment) Deserialize(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &s.MagicNumber); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.LittleEndian, &s.Version); err != nil {
		return err
	}
	if err := s.DocIDs.Deserialize(reader); err != nil {
		return fmt.Errorf("failed to deserialize DocIDs bitmap: %w", err)
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

// Serialize writes a block to the provided writer.
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

	numFreqs := uint32(len(b.TermFrequencies))
	if err := binary.Write(writer, binary.LittleEndian, numFreqs); err != nil {
		return fmt.Errorf("failed to write number of term frequencies: %w", err)
	}
	for _, freq := range b.TermFrequencies {
		if err := binary.Write(writer, binary.LittleEndian, freq); err != nil {
			return fmt.Errorf("failed to write term frequency: %w", err)
		}
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
	for i := uint32(0); i < numFreqs; i++ {
		var freq float32
		if err := binary.Read(reader, binary.LittleEndian, &freq); err != nil {
			return fmt.Errorf("failed to read term frequency: %w", err)
		}
		b.TermFrequencies[i] = freq
	}
	return nil
}

// WriteSegment writes a Segment to an io.Writer, typically a file.
func (s *Segment) WriteSegment(writer io.Writer) error {
	if err := s.Serialize(writer); err != nil {
		return fmt.Errorf("failed to serialize segment: %w", err)
	}
	return nil
}

// ReadSegment reads a Segment from an io.Reader, typically a file.
func (s *Segment) ReadSegment(reader io.Reader) error {
	if err := s.Deserialize(reader); err != nil {
		return fmt.Errorf("failed to deserialize segment: %w", err)
	}
	return nil
}
