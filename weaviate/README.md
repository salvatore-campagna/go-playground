# Storage Format Design for Weaviate Tech Challenge

This document presents a solution for the Weaviate Tech Challenge, which involves designing an efficient storage format for LSM-based inverted index segments. The solution focuses on efficient TF-IDF calculation and optimized disk storage without merging segments upfront.

## Context
The challenge requires designing an efficient storage format for inverted index segments, supporting:
- TF-IDF calculation across non-merged segments
- Operation on posting lists with millions of document IDs
- Efficient storage format

Inverted indexes are a core component of search engines, allowing for efficient retrieval of documents containing specific terms. Optimizing the storage of posting lists is critical to ensure both space efficiency and fast query performance.

## Initial Design Considerations

### DocID Handling:
While the sample data shows no docID overlaps between segments, I've implemented a base docID offset per segment (similar to what Lucene does) which:
- Handles potential overlaps if they exist allowing us to easily translate segment-local (unique) docIDs to globally unique docIDs
- Works efficiently if IDs are already unique
- Adds minimal overhead (one uint32 per segment)

### TF/IDF Calculation
For TF/IDF calculation, the idea is to aggregate document frequencies globally across segments. This approach ensures that the score reflects the true importance of terms relative to the entire dataset rather than just individual segments. Segments are just an underlying storage detail.

## Usage Pattern Analysis
According to Weaviate's documentation and architecture:
> "Filtered vector search in Weaviate is based on the concept of pre-filtering. This means that the filter is constructed before the vector search is performed." 
> "Pre-Filtering describes an approach where eligible candidates are determined before a vector search is started. The vector search then only considers candidates that are present on the 'allow' list"
> -- [Weaviate Documentation](https://weaviate.io/developers/weaviate/concepts/filtering)

As a result of this I made the following assumptions in the design of the storage format:
1. Filters are executed before returning the set of documents mathcing the filter
2. Support boolean operations between multiple filters
3. Compute term frequencies for any later scoring requirements

## Approaches to Efficient Storage

### Option 1: Simple Sorted Arrays
```
[Header]
[Sorted DocIDs Array]
[Term Frequencies Array]
```

Advantages:
- Implementation is straightforward and easy to understand
- Binary search can be used for efficient random lookups on Sorted DocIDs
- Direct array access makes term frequency retrieval fast
- Simple to maintain and debug
- Minimal memory overhead per posting list

Disadvantages:
- No compression means high storage costs for large document sets
- Boolean operations (AND, OR, NOT) require full array scans or merge operations
- Not space-efficient for sparse posting lists
- Large lists require loading complete arrays into memory
- Poor cache utilization when documents are far apart
- No optimization for dense vs sparse distributions

### Option 2: Delta-Encoded Arrays with Skip Lists
```
[Header]
[Skip List]
[Delta-Encoded DocIDs]
[Term Frequencies]
```

Advantages:
- Delta encoding good compression for sorted document IDs
- Skip lists enable faster searching and list intersection
- Good balance between random access and compression
- Works well with disk-based storage due to sequential reading patterns
- Can skip large portions of lists during boolean operations

Disadvantages:
- More complex implementation increases maintenance burden
- Skip list overhead can be significant for smaller lists
- Not as efficient as bitmap operations for dense sets
- Need careful tuning of skip list parameters
- Multiple data structures increase complexity
- Random access requires decoding multiple deltas

### Option 3: Roaring Bitmap-based Format (Chosen Solution)

After evaluating multiple storage formats, I chose a Roaring Bitmap-based approach due to its efficient handling of both sparse and dense data, fast boolean operations, and low memory footprint. This aligns well with Weaviate’s focus on pre-filtering before vector search.

### File Structure
The file is organized into segments, each containing blocks optimized for efficient retrieval and boolean operations.

#### Header Fields
- **Magic number (4 bytes)**: Identifies our custom format.
- **Version (1 byte)**: Ensures compatibility with future versions.
- **Term Length (2 bytes)**: Specifies the length of the term string.

#### Block Structure
Each block contains:
- **Min DocID**: The smallest DocID in the block.
- **Max DocID**: The largest DocID in the block.
- **Roaring Bitmap Container**: Stores document IDs in a compressed format.
- **Term Frequencies Array**: Stores term frequencies aligned with the bitmap.


```
[File Header]
  - Magic number (4 bytes): identifies our format
  - Version (1 byte): format version for compatibility with future version and backward compatibility
  - Term length (2 bytes): length of term string
  - Term string (variable): the actual term string
  - Base DocID (4 bytes): offset for this segment
  - Total documents (4 bytes): number of documents in segment
  - Number of blocks (4 bytes): how many blocks follow

[Blocks]
  [Block Header]
    - Min DocID (4 bytes): smallest DocID in block
    - Max DocID (4 bytes): largest DocID in block
    - Max TF (4 bytes float): highest term frequency in block
    - Block size (4 bytes): total bytes in block
  [Roaring Bitmap Container]
    - Container type (1 byte): array or bitmap
    - DocID storage: format depends on type:
      * Array: sorted list of 16-bit integers
      * Bitmap: 8KB bitmap (65536 bits)
  [Term Frequencies]
    - Array of float32: one per set bit in bitmap
    - Position determined by rank in bitmap
```

Example block containing documents with:
DocID:  100    103    105
TF:     0.3    0.5    0.2

```
[Block Header]
MinDocID: 100 (4 bytes)
MaxDocID: 105 (4 bytes)
MaxTF:    0.5 (4 bytes)
BlockSize: ... (4 bytes)    -> calculated depending on actual content

[Roaring Container]
Container Type: Array (1 byte) -> 0x01 (since < 4096 elements, 0x02 in case of >= 4096 elements)
Array Length: 3 (2 bytes)
Values (2 bytes each):         
  - 100
  - 103
  - 105

[Term Frequencies]
Array of float32 (4 bytes each):
  - 0.3
  - 0.5
  - 0.2
```

Advantages:
- Optimized for boolean operations which is crucial for filtering
- Adaptive storage format handles both sparse and dense data efficiently
- Proven in production systems (used by Elasticsearch, Lucene, etc.)
- Excellent compression while maintaining fast access
- CPU-efficient implementations available (SIMD instructions)
- Supports both iteration and random access patterns
- Natural fit for document filtering operations

Disadvantages:
- More complex than simple arrays requiring careful implementation
- Additional memory overhead for very small sets
- Need to maintain alignment between bitmap and term frequencies
- May be overkill for very small document sets

## Additional Design Choices

1. DocID Handling:
   - Each segment has a base DocID offset
   - Local DocIDs are mapped to global space using base offset
   - Enables independent processing of segments
   - Handles potential DocID overlaps between segments
   - Maintains ordering within segments

3. Block Organization:
   - Block metadata enables efficient skipping
   - Independent blocks allow partial loading

4. Key Features:
   - Block-based organization for memory efficint and efficient cache usage
   - Roaring bitmaps for efficient filtering operations
   - Aligned term frequencies for scoring calculations
   - Version field for format evolution and backward compatibility
   - Metadata favors optimizations allowing skipping entire blocks

## Rationale for Chosen Design

1. Alignment with Weaviate's Architecture:
   - Optimized for filtering-before-vector-search pattern
   - Efficient boolean operations for complex filters
   - Fast iteration over filtered document sets

2. Performance Characteristics:
   - Efficient boolean operations for complex filters
   - Good compression for both sparse and dense cases
   - Fast iteration over matching documents

3. Practical Considerations:
   - Industry-proven approach (similar to Elasticsearch/Lucene)
   - Good balance of complexity vs features
   - Room for optimization and extension

## Future Improvements

### Memory Efficiency

The currect design reads blocks in memory as they are needed but we could improve memory usage for instance
implementing memory mapping for segment files, with the following advantages:
- OS-level management of segments/blocks loaded in memory
- Readuced heap pressure since not all data needs to be in the application memory (heap)
- Sharing of segments/blocks between multiple processes

This becomes important when having:
- Multiple concurrent queries accessing the same segments/blocks
- Very large segments that do not fit in memory
- Optimized memory usage for systems with low memory compared to segment size

### Smarter Compression

Currently we store term frequencies as raw `floa32` values. Maybe we can implement better compression. Despite floating-point
compression being more challenging we might exploit integer-representation for floating-point values and apply integer compression
algorithms (delta encoding, GCD encoding, offset encoding, xor encoding PFopr, FastPFor,...)

### Adaptive Block Sizing

The currect fixed block size is simple but might not be optimal in scenarios like:
1. Dense regions with many documents might span multiple blocks while we could use a single larger block with a single bitmap container
2. Sparse regions with few documents might waste space for fixed-size blocks while we could just adapt the block size and maybe also changing data representation and compression techniques so that it works better for sparse data

### Format Evolution

The format will likely evolve over time and we would also need to support old segment/block formats for data that is already in production before new formats are introduced. As a result,
we need a strategy to deal with compatibility. As a result we might consider adding a further section to the segment or block header including something like:

```
[Header]
... existing fields ...
[Feature Flags]
- Compression flags or IDs
- Variable block size
- Custom metadata
```

This would enable us to:
- use different readers understanding different formats
- gradually rollout improvements and new features or optimizations
- deal with backward compatibility
- A/B test new optimizations by monitoring and comparing performance

The chosen Roaring Bitmap format provides an optimal balance between efficiency and complexity, aligning well with Weaviate’s architecture and requirements. Future improvements, such as adaptive block sizing and memory mapping, can further enhance performance, making this design scalable for production use.
