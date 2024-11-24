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
   - Uses **Roaring Bitmaps** for efficient document ID storage and iteration (inspired by the [Roaring Bitmaps paper](https://arxiv.org/pdf/1402.6407)).
   - Supports serialization and deserialization for persistence.
   - Includes configurable block size (via `MaxEntriesPerBlock`) to control how terms and posting lists are grouped into `Block`s.

2. **Query Engine (`engine`)**:
   - Processes multi-term queries across one or more segments.
   - Scores documents using **TF-IDF** to rank results based on relevance.
   - Implements block-level iterators and a min-heap for efficient processing.

3. **Testing Suite**:
   - Comprehensive unit tests ensure correctness of the scoring function, multi-segment queries, roaring bitmaps, and (some) edge cases.

4. **Fetcher (`fetcher`)**:
   - Reads and parses a JSON file with the format provided by the [Weaviate challenge dataset](https://storage.googleapis.com/weaviate-tech-challenges/db-engineer/segments.json).

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
Query: skywalker vader
Terms: [skywalker vader]

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
- **Min-Heap for Block Processing**:
  - Efficiently processes blocks by maintaining a priority queue of terms and their current positions, including DocID and term frequency.
- **TF-IDF Scoring**:
  - Dynamically computes scores as documents match all query terms.

#### **Trade-offs:**
- **Heap-based Query Processing**:
  - **Pros**: Reduces redundant scans and focuses on relevant document ranges.
  - **Cons**: Slightly complex to implement and debug due to iterator management.

---

## **5. Results**

The implementation fulfills all the requirements:
1. Efficient term-based indexing with Roaring Bitmaps and a custom binary file format.
2. Multi-term query execution using TF-IDF scoring.
3. Support for multi-segment queries using iterators for efficient data processing.

---

## **6. Next Steps**

1. **Performance Optimization**:
   - Parallelize query processing across multiple cores for faster execution.
2. **Index Compression**:
   - Explore advanced compression techniques for term frequencies.
3. **Scalability Testing**:
   - Benchmark performance on datasets with millions of documents.

---

See **TODOs** in the code for more details.

**NOTE:** This project incorporates concepts from my Roaring Bitmaps side project to learn Go, augmented with additional features to fulfill the challenge requirements.
