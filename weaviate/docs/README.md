# Weaviate Search Engine Coding Challenge

Author: Salvatore Campagna (Elasticsearch Storage Engine)

Email: salvatorecampagna@gmail.com  

---

This repository implements a **search engine** as part of a coding challenge for **Weaviate**, designed to process inverted index segments for efficient document retrieval and ranking. Below, I present the **problem breakdown**, **implementation details**, and the **design trade-offs** considered during the development of this solution.

---

## **1. Problem Statement**

The challenge was to implement a search engine that:
1. Indexes terms across documents and organizes them into segments using a suitable binary file format.
2. Supports multi-term queries using **TF-IDF scoring** to rank query results.
3. Handles large datasets with support for multiple segments and TF-IDF calculation.

---

## **2. Overview of the Solution**

### **Components**

1. **Storage Layer (`storage`)**:
   - Implements the core of the **inverted index**, storing terms and their associated documents.
   - Uses **Roaring Bitmaps** for efficient document ID storage and iteration (inspired by the paper [Better bitmap performance with Roaring bitmaps](https://arxiv.org/pdf/1402.6407)).
   - Supports serialization and deserialization for persistence.
   - Includes configurable block size (via `MaxEntriesPerBlock`) to control how terms and posting lists are grouped into `Block`s.

2. **Query Engine (`engine`)**:
   - Processes multi-term queries across one or more segments.
   - Scores documents using **TF-IDF** to rank results based on relevance.
   - Implements block-level iterators and a min-heap for efficient processing.

4. **Fetcher (`fetcher`)**:
   - Reads and parses a JSON file with the format provided by the [Weaviate challenge dataset](https://storage.googleapis.com/weaviate-tech-challenges/db-engineer/segments.json).

4. **Testing Suite**:
   - Comprehensive unit tests ensure correctness of the scoring function, multi-segment queries and roaring bitmaps, including edge cases.

5. **Command Line Utilities**:
   - Utilities to work with the index and query components. These are located in the `cmd` directory and can be executed via `make`. Key utilities include:
     - **Indexing (`index`)**: Builds an inverted index from a JSON file (`segments.json`) and stores it in `segment-data`.
     - **Querying (`query`)**: Executes queries against the index and retrieves matching documents.

---

## **3. Usage Instructions**

This section explains how to use the `index` and `query` commands to index documents and execute queries.

### **Prerequisites**
1. Install [Go](https://golang.org/dl/) if it's not already installed.
2. Ensure the dataset file `segments.json` is available. You can download it using:
   ```bash
   make download
   ```

### **Building the Project**
To build the indexing and querying utilities:
```bash
make build
```
This command compiles the source code and generates executables in the `bin/` directory:
- `bin/index` for indexing documents.
- `bin/query` for executing queries.

### **1. Indexing Documents**
To index documents from the dataset `segments.json`, run:
```bash
make index
```
This will:
1. Read the dataset from `segments.json`.
2. Create an inverted index.
3. Store the indexed data in the directory `segment-data/`.

#### **Custom Options**
You can specify custom paths for the dataset and the output directory:
```bash
bin/index -path=custom_segments.json -dir=custom_segment_data
```

### **2. Querying the Index**
To query the indexed documents, run:
```bash
make query
```
By default, this executes a query with the terms `great vector` (as specified in the code). Results are printed to the console.

#### **Custom Queries**
You can specify custom queries using the `-query` parameter:
```bash
bin/query -query="great database"
```

#### **Custom Index Directory**
If you stored the index in a custom directory, you can specify it using the `-dir` parameter:
```bash
bin/query -dir=custom_segment_data -query="great database"
```

#### **Example Output**
```plaintext
Query: "great vector"
Terms: ["great", "vector"]

Scored documents: 3
----------------------
| DocID    | Score   |
----------------------
| 5        | 3.14    |
| 10       | 2.71    |
| 15       | 2.30    |
----------------------
```

---

## **4. Implementation Details**

### **4.1 Storage Layer**

#### **Responsibilities:**
- Store terms and their corresponding posting lists (document IDs and term frequencies).
- Support fast access to posting lists for term-based lookups.
- Enable block-level organization for scalable query processing.

#### **Key Classes:**
- **`Segment`**: Represents a collection of indexed terms and their metadata.
- **`TermMetadata`**: Stores metadata for each term, including document frequency and associated blocks.
- **`Block`**: Encapsulates compressed document IDs and term frequencies using Roaring Bitmaps.

#### **Trade-offs:**
- **Roaring Bitmap**:
  - **Pros**: Minimizes memory usage while ensuring quick lookups.
  - **Cons**: Adds complexity to serialization and requires careful integration with scoring algorithms.
- **Block-Level Organization**:
  - **Pros**: Enables skipping blocks for queries with selective filters or rare terms, reducing redundant scans.
  - **Cons**: Adds overhead for queries requiring sequential access across all documents.

---

### **4.2 Query Engine**

#### **Responsibilities:**
- Execute multi-term queries across segments.
- Use **TF-IDF** scoring to compute relevance:
  - **TF (Term Frequency)**: Measures term importance in a document.
  - **IDF (Inverse Document Frequency)**: Reduces the weight of terms appearing in many documents.
- Ensure results are ranked and sorted based on scores.

#### **Algorithm Highlights**:

- **Coordinated traversal**:
  - Efficiently processes multiple iterators to find documents matching multiple terms and ranking them.
  - Each iterator processes a sorted list of document IDs (associated to a term and term frequency).
  - At any given time a min-heap is used to maintain the smalled `DocID` across all iterators.
  - If a document has all query terms the **TF-IDF** score is computed.
  - Only relevant documents are processed, while non-relevant ones are skipped
- **Min-Heap for Block Processing**:
  - Efficiently processes blocks by maintaining a priority queue of terms and their current positions, including DocID and term frequency.
- **TF-IDF Scoring**:
  - Dynamically computes scores as documents match all query terms.

#### **Trade-offs:**
- **Heap-based Query Processing**:
  - **Pros**: Reduces redundant scans and focuses on relevant document ranges.
  - **Cons**: Slightly complex to implement and debug due to iterator management.

---

### **5. Example Query Execution**

#### Query: `["rebels", "hope"]`

#### Blocks

- **Block 1**: `(term: "rebels", DocID: 1, tf: 1.0)`, `(term: "hope", DocID: 2, tf: 0.5)`
- **Block 2**: `(term: "empire", DocID: 1, tf: 1.0)`, `(term: "rebels", DocID: 2, tf: 0.7)`
- **Block 3**: `(term: "empire", DocID: 2, tf: 1.5)`, `(term: "hope", DocID: 3, tf: 1.2)`

---

#### Iterators and Heap at Each Step

##### Step 1

- `(term: "rebels", DocID: 1, tf: 1.0)`, `...`, `nil`
- `(term: "empire", DocID: 1, tf: 1.0)`, `...`, `nil`
- `(term: "empire", DocID: 2, tf: 1.5)`, `...`, `nil`

No match, discard `DocID:1`, as document with `DocID: 1` has terms "rebels" and "empire".

##### Step 2

- `(term: "hope",   DocID: 2, tf: 0.5)`, `nil`
- `(term: "rebels", DocID: 2, tf: 0.7)`, `nil`
- `(term: "empire", DocID: 2, tf: 1.5)`, `...`

Document with `DocID: 2` matches both "hope" and "rebels", as a result it is a match.

##### Score calculation

`DocID: 2` matches both "rebels" and "hope", the **TF-IDF** score is caldulated for document with `DocID: 2`.

- `DF(term: "rebels"): 2` (Documents `1` and `2`)
- `DF(term: "hope"): 2` (Documents `2` and `3`)
- `TF(term: "rebels", DocID: 2): 0.7`
- `DF(term: "rebels", DocID: 2): 0.5`
- `TF-IDF(term: "rebels", DocID: 2): 0.09`
- `TF-IDF(term: "hope", DocID: 2): 0.06`

The result is a score for document `DocID: 2` of `0.5`.

##### Step 3

Continues with no results.

---

#### Final Notes
- **Iterators**: One for each block, all managed in the heap.
- **Matching Documents**: Only `DocID: 2` matched all terms.

## **6. Results**

The implementation fulfills all the requirements:
1. Efficient term-based indexing with Roaring Bitmaps and a custom binary file format.
2. Multi-term query execution using TF-IDF scoring.
3. Support for multi-segment queries using iterators for efficient data processing.

---

## **7. Next Steps**

1. **Performance Optimization**:
   - Parallelize query processing across multiple cores for faster execution.
2. **Index Compression**:
   - Explore advanced compression techniques for term frequencies.
3. **Scalability Testing**:
   - Benchmark performance on datasets with millions of documents.

---

See **TODOs** in the code for more details.