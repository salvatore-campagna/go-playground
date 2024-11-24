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
package engine
