// Package storage provides efficient data structures and algorithms for managing
// term-document relationships in an inverted index. It is designed for use in
// search engines and inverted index implementations.
//
// # Overview
//
// The storage package combines several key components to enable efficient
// document retrieval and indexing. It leverages compressed data structures such as
// Roaring Bitmaps for memory-efficient representation of document IDs and provides
// tools for managing posting lists, term frequencies, and block-level metadata.
//
// # Features
//
// - **Bitmap Iterators**: Enable iteration over Roaring Bitmap containers for document IDs.
// - **Posting List Iterators**: Facilitate traversal of posting lists with term frequencies.
// - **Block-Level Access**: Allow efficient sequential access to blocks of posting lists.
// - **Roaring Bitmap Containers**: Support sparse and dense data sets with Array and Bitmap containers.
// - **Set Operations**: Include union and intersection for advanced query handling.
// - **Serialization**: Provide support for saving and loading storage components.
//
// # Roaring Bitmaps
//
// The package uses Roaring Bitmaps for compact and high-performance representation of
// document IDs. Roaring Bitmaps dynamically adjust to sparse and dense data distributions
// using the following container types:
//
// - **ArrayContainer**: For sparse sets of integers, stores values as a sorted array of `uint16`.
// - **BitmapContainer**: For dense sets of integers, uses a set of `uint16` words.
//
// These containers enable efficient operations such as unions and intersections, making them
// ideal for query processing in search engines.
//
// ## Example Use Case
//
// Consider a bitmap index storing document IDs for terms in a search engine. Each term is associated with
// a bitmap, where the presence of a document ID indicates that the term appears in that document.
// Operations like union and intersection enable powerful queries, such as:
//   - Find all documents containing any of a set of terms (union).
//   - Find all documents containing all of a set of terms (intersection).
//   - Exclude documents marked as deleted using a difference operation (TODO).
//
// # Iterators
//
// Iterators are core to the `storage` package, enabling traversal of posting lists and
// Roaring Bitmap containers. These iterators are designed for high performance, with
// optimizations for sequential access and skipping over non-relevant document IDs.
//
// ## Types of Iterators
//
//   - **Bitmap Iterator**: Traverses document IDs stored in a Roaring Bitmap. Supports both
//     forward iteration and batch-based retrieval for scalability.
//   - **Posting List Iterator**: Combines document IDs with term frequencies. Useful for
//     ranking and scoring during query processing.
//
// ## Block-Level Access
//
// Posting lists are organized into blocks for efficient access. Each block contains:
// - A compressed representation of document IDs (using Roaring Bitmaps).
// - Term frequencies associated with the document IDs.
// - Metadata such as minimum and maximum document IDs in the block.
//
// Iterators support skipping over blocks based on query constraints, allowing faster traversal.
//
// # File Format
//
// The segment file format is organized into three main sections:
//
// ## File Header
// The file header provides metadata for the segment:
//   - Magic Number (4 bytes): Identifies the segment file format (e.g., 0x007E8B11)
//   - Version (1 byte): Segment format version
//   - Document Count (4 bytes): Total number of documents
//   - Number of Terms (4 bytes): Total number of unique terms
//
// ## Terms Section
// Each term has metadata and associated posting list blocks:
//   - Term Length (2 bytes): Length of the term string
//   - Term String (variable): UTF-8 encoded term
//   - Total Documents (4 bytes): Number of documents containing the term
//   - Number of Blocks (4 bytes): Number of blocks for the term
//
// ## Blocks Section
// Each block represents a chunk of the posting list:
//   - Min DocID (4 bytes): Minimum document ID in the block
//   - Max DocID (4 bytes): Maximum document ID in the block (unused at the moment)
//   - Bitmap Type (1 byte): Type of Roaring Bitmap container (1 = Array, 2 = Bitmap)
//   - DocIDs: Compressed Roaring Bitmap for document IDs
//   - Term Frequencies ([]float32): Frequencies of the term in each document
//
// ## Example Layout
//
// The following example illustrates a segment file:
//
// Magic Number        : 0x007E8B11
// Version             : 1
// Document Count      : 2500
// Number of Terms     : 2
//
// **[Term 1: "database"]**
//   - Total Documents   : 2000
//   - Number of Blocks  : 2
//   - **Block 1**
//   - Min DocID      : 1
//   - Max DocID      : 1000
//   - Bitmap Type    : ArrayContainer
//   - DocIDs         : [1, 50, ..., 1000]
//   - Term Frequencies: [0.5, 0.6, ..., 0.8]
//   - **Block 2**
//   - Min DocID      : 1001
//   - Max DocID      : 2000
//   - Bitmap Type    : BitmapContainer
//   - DocIDs         : [1001, 1050, ..., 2000]
//   - Term Frequencies: [0.2, 0.4, ..., 0.9]
//
// **[Term 2: "vector"]**
//   - Total Documents   : 500
//   - Number of Blocks  : 1
//   - **Block 1**
//   - Min DocID      : 3000
//   - Max DocID      : 3500
//   - Bitmap Type    : BitmapContainer
//   - DocIDs         : [3000, 3050, ..., 3500]
//   - Term Frequencies: [0.1, 0.3, ..., 0.7]
package storage
