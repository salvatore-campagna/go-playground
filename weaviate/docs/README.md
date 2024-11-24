# Full-Text Search Engine Coding Challenge

This repository implements a **full-text search engine** designed to process inverted index segments for efficient document retrieval and ranking. Below, I present the **problem breakdown**, **implementation details**, and the **design trade-offs** I considered during the development of this solution.

---

## **1. Problem Statement**

The challenge was to implement a search engine that:
1. Indexes terms across documents and organizes them into segments.
2. Supports multi-term queries using **TF-IDF scoring** to rank results.
3. Handles large datasets supporting for multiple segments and TF-IDF calculation.

---

## **2. Overview of the Solution**

### **Components**

1. **Storage Layer (`storage`)**:
   - Implements the core of the **inverted index**, storing terms and their associated documents.
   - Uses **Roaring Bitmaps** for efficient document ID storage and iteration (inspired by [Roaring Bitmaps paper](https://arxiv.org/pdf/1402.6407)).
   - Supports serialization and deserialization for persistence.
   - Configurable block size (see `MaxEntriesPerBlock`) controls how terms and posting lists are grouped into `Block`s.

2. **Query Engine (`engine`)**:
   - Processes multi-term queries across one or more segments.
   - Scores documents using **TF-IDF** to rank results based on relevance.
   - Implements block-level iterators and a min-heap for efficient processing.

3. **Testing Suite**:
   - Comprehensive unit tests ensure correctness of the scoring function, multi-segment queries, and edge cases.

4. **Fetcher (`fetcher`)**:
   - Reads and parses a JSON file with the format provided by the [Weaviate challenge dataset](https://storage.googleapis.com/weaviate-tech-challenges/db-engineer/segments.json).

5. **Command Line Utilities**
   - Utilities to work with the index and query components. These are located in the `cmd` directory and can be executed via `make`. Key utilities include:
     - **Indexing (`index`)**: Builds an inverted index from a JSON file (`segments.json`) and stores it in `segment-data`.
     - **Querying (`query`)**: Executes queries against the index and retrieves matching documents.
     - **Data Generation (`datagen`)**: Generates JSON test data with the same format as the Weaviate challenge dataset](https://storage.googleapis.com/weaviate-tech-challenges/db-engineer/segments.json).
     - **Statistics (`stats`)**: Computes statistics (e.g., term frequency, document distribution).
     - **Data Cleaning (`dataclean`)**: Cleans and preprocesses JSON input before indexing.

   - See the **Makefile Usage** section for details.

---

## **3. Implementation Details**

### **3.1 Storage Layer**

#### **Responsibilities:**
- Store terms and their corresponding posting lists (documents and term frequencies).
- Support fast access to posting lists for term-based lookups.
- Enable block-level organization for scalable query processing.

#### **Key Classes:**
- **`Segment`**: Represents a collection of indexed terms and their metadata.
- **`TermMetadata`**: Stores metadata for each term, including document frequency and associated blocks.
- **`Block`**: Encapsulates compressed document IDs and term frequencies using Roaring Bitmaps.

#### **Trade-offs:**
- **Roaring Bitmap**: Chosen for efficient compression and fast iteration over document IDs.
  - **Pros**: Minimizes memory usage while ensuring quick lookups.
  - **Cons**: Adds complexity to serialization and requires careful integration with scoring algorithms.
- **Block-Level Organization**: Facilitates efficient query processing by allowing skipping over irrelevant document ranges (see `MinDocID` and `MaxDocID`).

---

### **3.2 Query Engine**

#### **Responsibilities:**
- Execute multi-term queries across segments.
- Use **TF-IDF** scoring to compute relevance:
  - **TF (Term Frequency)**: Measures term importance in a document.
  - **IDF (Inverse Document Frequency)**: Reduces the weight of terms appearing in many documents.
- Ensure results are ranked and sorted based on scores.

#### **Algorithm Highlights**:
- **Min-Heap for Block Processing**:
  - Efficiently processes blocks by maintaining a priority queue of terms and their current positions.
- **TF-IDF Scoring**:
  - Dynamically computes scores as documents match all query terms.

#### **Key Challenges:**
- Correct calculation of IDF (`log((totalDocs + 1) / (docFrequency + 1))`) to avoid division by zero or inaccurate scores.
- Handling multi-segment queries, ensuring documents across segments contribute to the final results.

#### **Trade-offs:**
- **Heap-based Query Processing**:
  - **Pros**: Efficient for large datasets by reducing redundant scans.
  - **Cons**: Slightly complex to implement and debug due to iterator (correct) management.

---

### **3.3 Testing**

#### **Goals**:
- Validate the correctness of indexing, scoring, and query execution.
- Test both edge cases and real-world-inspired scenarios.

#### **Key Tests**:
1. **Single-Term Query**:
   - Ensures documents matching a single term are ranked correctly.
2. **Multi-Term Query**:
   - Validates that documents matching all terms are retrieved and scored appropriately.
3. **Multi-Segment Query**:
   - Tests queries across multiple segments, verifying that scores are computed consistently.

---

## **4. Trade-offs and Design Decisions**

### **4.1 Roaring Bitmaps**
- **Why?**: Provides fast document ID retrieval while minimizing storage overhead.
- **Trade-off**: Added complexity in integrating bitmap operations with scoring and block iterators.

### **4.2 TF-IDF Scoring**
- **Why?**: Standard metric for information retrieval, balancing term frequency with document relevance.
- **Improvement**: Dynamically computed scores avoid precomputing and storing values, reducing memory requirements.

### **4.3 Min-Heap for Query Processing**
- **Why?**: Optimizes traversal of posting lists and block-level skipping.
- **Trade-off**: Slightly increased runtime overhead for heap management.

### **4.4 Multi-Term Query**
- **Why?**: Prioritizes precision by returning documents that match all queried terms, ensuring high relevance.
- **Trade-off**: May reduce recall by excluding partially relevant documents, which could be limiting for broader searches.

---

## **5. Key Challenges**

1. **IDF Calculation**:
   - Ensuring correct handling of edge cases, such as low document frequencies, to avoid zero or negative scores.

2. **Multi-Segment Queries**:
   - Ensuring consistent results when documents are spread across multiple segments.

3. **Testing**:
   - Balancing between synthetic test cases for edge scenarios and real-world-inspired tests for usability.

---

## **6. Multi Query Execution Example**

To demonstrate how the query engine processes multi-term queries using blocks, consider the following setup:

### **Query Scenario**

Suppose we are querying for documents that contain both the terms **"skywalker"** and **"vader"**. The index is organized into blocks as follows:

- **Term: skywalker**
  - **Block 1:** {(docID: 1, tf: 3), (docID: 5, tf: 2), (docID: 10, tf: 1)}
  - **Block 2:** {(docID: 15, tf: 4), (docID: 20, tf: 3)}

- **Term: vader**
  - **Block 1:** {(docID: 5, tf: 3), (docID: 8, tf: 2), (docID: 10, tf: 1)}
  - **Block 2:** {(docID: 15, tf: 3), (docID: 22, tf: 2)}

### **Step-by-Step Query Execution**

#### **1. Initialization**
- Create iterators for each block:
  - **skywalker Block 1 Iterator:** Starts at (docID: 1, tf: 3).
  - **skywalker Block 2 Iterator:** Starts at (docID: 15, tf: 4).
  - **vader Block 1 Iterator:** Starts at (docID: 5, tf: 3).
  - **vader Block 2 Iterator:** Starts at (docID: 15, tf: 3).

- Initialize the **min-heap** with the smallest `docID` from each iterator:
    - Heap = [(skywalker, Block 1, docID 1), (vader, Block 1, docID 5), (skywalker, Block 2, docID 15), (vader, Block 2, docID 15)]

#### **2. Process docID 1**
- Pop the smallest entry: `(skywalker, Block 1, docID 1)`.
- Process `docID 1` for **skywalker**.
- Advance the **skywalker Block 1 Iterator** to `docID 5, tf: 2` and push it back into the heap:
    - Heap = [(vader, Block 1, docID 5), (skywalker, Block 1, docID 5), (skywalker, Block 2, docID 15), (vader, Block 2, docID 15)]


#### **3. Process docID 5**
- Pop `(vader, Block 1, docID 5)` and `(skywalker, Block 1, docID 5)`.
- Both terms match at `docID 5`. Compute the **TF-IDF score**:
- **skywalker (tf: 2)**, **vader (tf: 3)**.
- `Score(docID 5) = TF-IDF(skywalker) + TF-IDF(vader)`.

#### **4. Process docID 10**
- Process `docID 10` for both terms.
- Compute the **TF-IDF score** for `docID 10`:
- **skywalker (tf: 1)**, **vader (tf: 1)**.
- `Score(docID 10) = TF-IDF(skywalker) + TF-IDF(vader)`.

#### **5. Process docID 15**
- Process `docID 15` for both terms from **skywalker Block 2** and **vader Block 2**.
- Compute the **TF-IDF score** for `docID 15`:
- **skywalker (tf: 4)**, **vader (tf: 3)**.
- `Score(docID 15) = TF-IDF(skywalker) + TF-IDF(vader)`.

#### **6. Process Remaining Entries**
- Process remaining `docIDs` (e.g., `docID: 22` for **vader**) until all iterators are exhausted.

### **Final Results**
The query retrieves three matching documents:
1. **docID: 5** (Matches both terms in **skywalker Block 1** and **vader Block 1**).
2. **docID: 10** (Matches both terms in **skywalker Block 1** and **vader Block 1**).
3. **docID: 15** (Matches both terms in **skywalker Block 2** and **vader Block 2**).

Each document is scored using its **TF-IDF** values for the terms.

---


## **7. Results**

The implementation fulfills all the requirements:
1. Efficient term-based indexing with Roaring Bitmaps and custom binary file format.
2. Multi-term query execution using TF-IDF scoring.
3. Support for multi-segment queries using iterators for efficient data processing.

---

## **8. Next Steps**

1. **Performance Optimization**:
   - Parallelize query processing across multiple cores for faster execution.
2. **Index Compression**:
   - Explore advanced compression techniques for term frequencies.
3. **Scalability Testing**:
   - Benchmark performance on datasets with millions of documents.

See **TODOs** in the code for more datails.
