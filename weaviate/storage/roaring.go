// Package storage implements a memory-efficient compressed bitmap index using the Roaring Bitmap format.
// It provides optimized containers for both sparse and dense data sets, supporting fast set operations
// like unions and intersections. The implementation follows the Roaring Bitmap specification detailed
// at https://roaringbitmap.org/.
package storage

// TODO: Use Rank method to access terms effeciently
// TODO: Implement Run-Length Encoding (RLE) container
// TODO: Implement varint encoding for container metadata
// TODO: Add versioning support for serialization format
// TODO: Evaluate implementing SIMD operations for bitmap operations (Go Assembly)
// TODO: Add container-level parallel prcoessing
// TODO: Implement bulk Add operation
// TODO: Do we need support for concurrent access?
// TODO: Consider using Snappy or Zstandard to comrpess serialized data
// TODO: Use custom errors instead fmt.Errorf
// TODO: Benchmarking and Profiling to guide optimizations
// TODO: Support for more operations like NOT or XOR or DIFF
// TODO: Add checksums to check for data integrity
// TODO: Explore other compression mechanisms
// TODO: Implement Diff inc ase we need to support DELETE operations on documents
// TODO: Evaluate if caching cardinality makes sense (worse Union/Intersection)

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"sort"
	"weaviate/encoders"
)

const ContainerConversionThreshold = 4096

// ContainerType identifies the internal container implementation.
type ContainerType uint8

const (
	ArrayContainerType ContainerType = iota + 1
	BitmapContainerType
)

// RoaringContainer defines the interface for bitmap storage containers.
// Implementations must support basic set operations and serialization.
type RoaringContainer interface {
	Add(value uint16)
	Contains(value uint16) bool
	Cardinality() int
	Union(other RoaringContainer) RoaringContainer
	Intersection(other RoaringContainer) RoaringContainer
	Serialize(io.Writer) error
	Deserialize(io.Reader) error
}

// ArrayContainer implements RoaringContainer using a sorted array,
// optimized for sparse data sets (cardinality < 4096).
type ArrayContainer struct {
	values      []uint16
	cardinality int
	encoder     encoders.ArrayEncoderDecoder
}

// TODO: smarter encoder configuration
// TODO: chose a better value for `minLen` 128 for the delta encoder below
// NewArrayContainer creates an empty ArrayContainer.
// Array containers are delta encoded (they store sorted integers)
func NewArrayContainer() *ArrayContainer {
	return &ArrayContainer{
		values:      []uint16{},
		cardinality: 0,
		encoder:     encoders.NewPlainEncoder(),
	}
}

// Add inserts a value into the ArrayContainer maintaining sort order.
// If the value already exists, the container remains unchanged.
func (ac *ArrayContainer) Add(value uint16) {
	// TODO Insertion with shifting is inefficient
	index := sort.Search(len(ac.values), func(i int) bool { return ac.values[i] >= value })
	if index < len(ac.values) && ac.values[index] == value {
		return
	}
	ac.values = append(ac.values, 0)
	copy(ac.values[index+1:], ac.values[index:])
	ac.values[index] = value
	ac.cardinality++
}

// Contains checks if a value exists in the ArrayContainer using binary search.
func (ac *ArrayContainer) Contains(value uint16) bool {
	index := sort.Search(len(ac.values), func(i int) bool { return ac.values[i] >= value })
	return index < len(ac.values) && ac.values[index] == value
}

// Cardinality returns the number of unique values in the container.
func (ac *ArrayContainer) Cardinality() int {
	return ac.cardinality
}

// Serialize writes the ArrayContainer's data to the provided writer in a compact format.
// Serialize writes the ArrayContainer's data to the provided writer in a compact format.
func (ac *ArrayContainer) Serialize(writer io.Writer) error {
	length := uint16(len(ac.values))
	if err := binary.Write(writer, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("error while serializing array container length: %v", err)
	}

	if err := ac.encoder.Encode(ac.values, writer); err != nil {
		return fmt.Errorf("error while encoding array container: %v", err)
	}

	return nil
}

// Deserialize reads ArrayContainer data from the provided reader.
func (ac *ArrayContainer) Deserialize(reader io.Reader) error {
	var length uint16
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return fmt.Errorf("error while deserializing array container length: %v", err)
	}
	decodedValues, err := ac.encoder.Decode(reader, int(length))
	if err != nil {
		return fmt.Errorf("error while decoding array container: %v", err)
	}

	if len(decodedValues) != int(length) {
		return fmt.Errorf("error while decoding array container, expected len %d, got len %d", length, len(decodedValues))
	}

	ac.values = decodedValues
	ac.cardinality = len(decodedValues)
	return nil
}

// Rank returns the number of values less than or equal to the input value.
func (ac *ArrayContainer) Rank(value uint16) int {
	return sort.Search(len(ac.values), func(i int) bool { return ac.values[i] > value })
}

// Union combines two containers and returns a new container with all unique values.
func (ac *ArrayContainer) Union(other RoaringContainer) RoaringContainer {
	switch other := other.(type) {
	case *ArrayContainer:
		result := NewArrayContainer()
		i, j := 0, 0
		for i < len(ac.values) && j < len(other.values) {
			if ac.values[i] < other.values[j] {
				result.Add(ac.values[i])
				i++
			} else if ac.values[i] > other.values[j] {
				result.Add(other.values[j])
				j++
			} else {
				result.Add(ac.values[i])
				i++
				j++
			}
		}

		for i < len(ac.values) {
			result.Add(ac.values[i])
			i++
		}

		for j < len(other.values) {
			result.Add(other.values[j])
			j++
		}
		return result

	case *BitmapContainer:
		return other.Union(ac)
	}
	return nil
}

// Intersection returns a new container with values present in both containers.
func (ac *ArrayContainer) Intersection(other RoaringContainer) RoaringContainer {
	switch other := other.(type) {
	case *ArrayContainer:
		result := NewArrayContainer()
		i, j := 0, 0
		for i < len(ac.values) && j < len(other.values) {
			if ac.values[i] < other.values[j] {
				i++
			} else if ac.values[i] > other.values[j] {
				j++
			} else {
				result.Add(ac.values[i])
				i++
				j++
			}
		}
		return result

	case *BitmapContainer:
		return other.Intersection(ac)
	}
	return nil
}

// ToBitmapContainer converts an ArrayContainer to a BitmapContainer.
// This is typically called when the container grows beyond the conversion threshold.
func (ac *ArrayContainer) ToBitmapContainer() *BitmapContainer {
	bc := NewBitmapContainer()
	for _, value := range ac.values {
		bc.Add(value)
	}
	return bc
}

// BitmapContainer implements RoaringContainer using a bitmap,
// optimized for dense data sets (cardinality > 4096).
type BitmapContainer struct {
	Bitmap      []uint64
	cardinality int
}

// NewBitmapContainer creates a fixed-size BitmapContainer to handle all possible uint16 values.
func NewBitmapContainer() *BitmapContainer {
	return &BitmapContainer{
		Bitmap:      make([]uint64, 1024), // Preallocate for 65536 bits (1024 * 64 bits)
		cardinality: 0,
	}
}

// Add sets the bit corresponding to the value in the bitmap.
func (bc *BitmapContainer) Add(value uint16) {
	word := int(value / 64) // Calculate the word index
	bit := uint(value % 64) // Calculate the bit position within the word

	// Dynamically expand the bitmap if the word index exceeds the current capacity
	if word >= len(bc.Bitmap) {
		newBitmap := make([]uint64, word+1)
		copy(newBitmap, bc.Bitmap)
		bc.Bitmap = newBitmap
	}

	// Check if the bit is already set, if not, set it and increment the cardinality
	if (bc.Bitmap[word] & (1 << bit)) == 0 {
		bc.Bitmap[word] |= (1 << bit)
		bc.cardinality++
	}
}

// Contains checks if the bit corresponding to the value is set in the bitmap.
func (bc *BitmapContainer) Contains(value uint16) bool {
	word := value / 64
	bit := value % 64
	return (bc.Bitmap[word] & (1 << bit)) != 0
}

// Cardinality returns the number of bits set in the bitmap.
func (bc *BitmapContainer) Cardinality() int {
	return bc.cardinality
}

// Serialize writes the BitmapContainer's data to the provided writer.
func (bc *BitmapContainer) Serialize(writer io.Writer) error {
	length := uint32(len(bc.Bitmap))
	if err := binary.Write(writer, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("error while serializing bitmap container length: %v", err)
	}

	for i := 0; i < int(length); i++ {
		if err := binary.Write(writer, binary.LittleEndian, bc.Bitmap[i]); err != nil {
			return fmt.Errorf("error while serializing bitmap container: %v", err)
		}
	}

	if err := binary.Write(writer, binary.LittleEndian, uint32(bc.cardinality)); err != nil {
		return fmt.Errorf("error while serializing bitmap container cardinality: %v", err)
	}
	return nil
}

// Deserialize reads BitmapContainer data from the provided reader.
func (bc *BitmapContainer) Deserialize(reader io.Reader) error {
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return fmt.Errorf("error while deserializing bitmap container length: %v", err)
	}

	bc.Bitmap = make([]uint64, length)

	for i := 0; i < int(length); i++ {
		if err := binary.Read(reader, binary.LittleEndian, &bc.Bitmap[i]); err != nil {
			return fmt.Errorf("error while deserializing bitmap container: %v", err)
		}
	}

	var cardinality uint32
	if err := binary.Read(reader, binary.LittleEndian, &cardinality); err != nil {
		return fmt.Errorf("error while deserializing bitmap container cardinality: %v", err)
	}
	bc.cardinality = 0
	for _, word := range bc.Bitmap {
		bc.cardinality += bits.OnesCount64(word)
	}

	if uint32(bc.cardinality) != cardinality {
		return fmt.Errorf("cardinality mismatch in bitmap container, expected: %d, got: %d", cardinality, bc.cardinality)
	}
	return nil
}

// Union performs a bitwise OR operation between two containers.
func (bc *BitmapContainer) Union(other RoaringContainer) RoaringContainer {
	switch other := other.(type) {
	case *BitmapContainer:
		result := NewBitmapContainer()
		for i := 0; i < len(bc.Bitmap); i++ {
			result.Bitmap[i] = bc.Bitmap[i] | other.Bitmap[i]
		}
		result.cardinality = 0
		for _, word := range result.Bitmap {
			result.cardinality += bits.OnesCount64(word)
		}
		return result

	case *ArrayContainer:
		return bc.Union(other.ToBitmapContainer())
	}
	return nil
}

// Intersection performs a bitwise AND operation between two containers.
func (bc *BitmapContainer) Intersection(other RoaringContainer) RoaringContainer {
	switch other := other.(type) {
	case *BitmapContainer:
		result := NewBitmapContainer()
		for i := 0; i < len(bc.Bitmap); i++ {
			result.Bitmap[i] = bc.Bitmap[i] & other.Bitmap[i]
		}
		result.cardinality = 0
		for _, word := range result.Bitmap {
			result.cardinality += bits.OnesCount64(word)
		}
		return result

	case *ArrayContainer:
		result := NewArrayContainer()
		for _, v := range other.values {
			if bc.Contains(v) {
				result.Add(v)
			}
		}
		return result
	}
	return nil
}

// Rank returns the number of bits set before or at the given value.
func (bc *BitmapContainer) Rank(value uint16) int {
	wordIndex := int(value / 64)
	bitPosition := int(value % 64)

	if wordIndex >= len(bc.Bitmap) {
		return bc.Cardinality()
	}

	rank := 0
	for i := 0; i < wordIndex; i++ {
		rank += bits.OnesCount64(bc.Bitmap[i])
	}

	mask := (uint64(1) << (bitPosition + 1)) - 1
	rank += bits.OnesCount64(bc.Bitmap[wordIndex] & mask)

	return rank
}

// ToArrayContainer converts a BitmapContainer to an ArrayContainer.
// This is typically called when the container becomes sparse.
func (bc *BitmapContainer) ToArrayContainer() *ArrayContainer {
	ac := NewArrayContainer()
	for i, word := range bc.Bitmap {
		if word != 0 {
			for bit := 0; bit < 64; bit++ {
				if (word & (1 << bit)) != 0 {
					ac.Add(uint16(i*64 + bit))
				}
			}
		}
	}
	return ac
}

// RoaringBitmap implements a compressed bitmap using a two-level indexing structure.
// The first level splits values on the high 16 bits, mapping them to optimized containers
// storing the low 16 bits.
type RoaringBitmap struct {
	containers  map[uint16]RoaringContainer
	cardinality int
}

// NewRoaringBitmap creates an empty RoaringBitmap.
func NewRoaringBitmap() *RoaringBitmap {
	return &RoaringBitmap{
		containers:  make(map[uint16]RoaringContainer),
		cardinality: 0,
	}
}

// Add inserts a value into the appropriate container, creating a new container if necessary.
// Automatically converts ArrayContainers to BitmapContainers when they exceed the threshold.
// Add inserts a value into the appropriate container, creating a new container if necessary.
// Automatically converts ArrayContainers to BitmapContainers when they exceed the threshold.
func (rb *RoaringBitmap) Add(value uint32) {
	key := uint16(value >> 16)    // Extract the high-order 16 bits
	low := uint16(value & 0xFFFF) // Extract the low-order 16 bits

	container, exists := rb.containers[key]
	if !exists {
		container = NewArrayContainer()
		rb.containers[key] = container
	}

	initialCardinality := container.Cardinality()
	container.Add(low)
	if container.Cardinality() > initialCardinality {
		rb.cardinality++
	}

	if ac, ok := container.(*ArrayContainer); ok && ac.Cardinality() > ContainerConversionThreshold {
		rb.containers[key] = ac.ToBitmapContainer()
	}
}

// Contains checks if a value exists in the bitmap.
func (rb *RoaringBitmap) Contains(value uint32) bool {
	key := uint16(value >> 16)
	low := uint16(value & 0xFFFF)

	container, exists := rb.containers[key]
	if !exists {
		return false
	}
	return container.Contains(low)
}

// Union returns a new bitmap containing all values from both bitmaps.
func (rb *RoaringBitmap) Union(other *RoaringBitmap) *RoaringBitmap {
	result := NewRoaringBitmap()
	result.cardinality = 0

	for key, container := range rb.containers {
		result.containers[key] = container
		result.cardinality += container.Cardinality()
	}

	for key, container := range other.containers {
		if existing, exists := result.containers[key]; exists {
			unionContainer := existing.Union(container)
			result.containers[key] = unionContainer
			result.cardinality += unionContainer.Cardinality() - existing.Cardinality()
		} else {
			result.containers[key] = container
			result.cardinality += container.Cardinality()
		}
	}

	return result
}

// Intersection returns a new bitmap containing values present in both bitmaps.
func (rb *RoaringBitmap) Intersection(other *RoaringBitmap) *RoaringBitmap {
	result := NewRoaringBitmap()
	result.cardinality = 0

	for key, container := range rb.containers {
		if otherContainer, exists := other.containers[key]; exists {
			intersectionContainer := container.Intersection(otherContainer)
			if intersectionContainer.Cardinality() > 0 {
				result.containers[key] = intersectionContainer
				result.cardinality += intersectionContainer.Cardinality()
			}
		}
	}

	return result
}

// Cardinality returns the total number of unique values in the bitmap.
func (rb *RoaringBitmap) Cardinality() int {
	return rb.cardinality
}

// Serialize writes the RoaringBitmap to the provided writer in a portable format.
func (rb *RoaringBitmap) Serialize(writer io.Writer) error {
	numContainers := uint32(len(rb.containers))
	if err := binary.Write(writer, binary.LittleEndian, numContainers); err != nil {
		return fmt.Errorf("failed to write number of containers: %w", err)
	}

	for key, container := range rb.containers {
		if err := binary.Write(writer, binary.LittleEndian, key); err != nil {
			return fmt.Errorf("failed to write container key: %w", err)
		}

		var containerType ContainerType
		switch container.(type) {
		case *ArrayContainer:
			containerType = ArrayContainerType
		case *BitmapContainer:
			containerType = BitmapContainerType
		default:
			return fmt.Errorf("unknown container type: %T", container)
		}

		if err := binary.Write(writer, binary.LittleEndian, containerType); err != nil {
			return fmt.Errorf("failed to write container type: %w", err)
		}

		if err := container.Serialize(writer); err != nil {
			return fmt.Errorf("failed to serialize container: %w", err)
		}
	}

	return nil
}

// Deserialize reads a previously serialized RoaringBitmap from the provided reader.
func (rb *RoaringBitmap) Deserialize(reader io.Reader) error {
	rb.containers = make(map[uint16]RoaringContainer)

	var numContainers uint32
	if err := binary.Read(reader, binary.LittleEndian, &numContainers); err != nil {
		return fmt.Errorf("failed to read number of containers: %w", err)
	}

	for i := uint32(0); i < numContainers; i++ {
		var key uint16
		if err := binary.Read(reader, binary.LittleEndian, &key); err != nil {
			return fmt.Errorf("failed to read container key: %w", err)
		}

		var containerType ContainerType
		if err := binary.Read(reader, binary.LittleEndian, &containerType); err != nil {
			return fmt.Errorf("failed to read container type: %w", err)
		}

		var container RoaringContainer
		switch containerType {
		case ArrayContainerType:
			container = NewArrayContainer()
		case BitmapContainerType:
			container = NewBitmapContainer()
		default:
			return fmt.Errorf("unknown container type: %d", containerType)
		}

		if err := container.Deserialize(reader); err != nil {
			return fmt.Errorf("failed to deserialize container data: %w", err)
		}

		rb.containers[key] = container
	}

	rb.cardinality = 0
	for _, container := range rb.containers {
		rb.cardinality += container.Cardinality()
	}

	return nil
}

// TODO: replace with iterator to use less memory and allow early stopping
func (rb *RoaringBitmap) GetDocIDs() []uint32 {
	var docIDs []uint32
	for key, container := range rb.containers {
		base := uint32(key) << 16
		switch c := container.(type) {
		case *ArrayContainer:
			for _, val := range c.values {
				docIDs = append(docIDs, base|uint32(val))
			}
		case *BitmapContainer:
			for i, word := range c.Bitmap {
				if word == 0 {
					continue
				}
				for bit := 0; bit < 64; bit++ {
					if word&(1<<bit) != 0 {
						docID := base | uint32(i*64+bit)
						docIDs = append(docIDs, docID)
					}
				}
			}
		}
	}
	return docIDs
}

// Rank counts the number of values (docIDs) in the RoaringBitmap up to the given value.
func (rb *RoaringBitmap) Rank(docId uint32) (int, error) {
	rank := 0
	targetKey := uint16(docId >> 16)
	targetLow := uint16(docId & 0xFFFF)
	for key, container := range rb.containers {
		if key < targetKey {
			rank += container.Cardinality()
		} else if key == targetKey {
			switch container := container.(type) {
			case *ArrayContainer:
				rank += container.Rank(targetLow)
			case *BitmapContainer:
				rank += container.Rank(targetLow)
			default:
				return 0, fmt.Errorf("unknown container")
			}
		}
	}

	return rank, nil
}
