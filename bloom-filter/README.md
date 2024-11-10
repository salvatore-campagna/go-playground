# Bloom Filter Exercise

## Introduction

A **Bloom Filter** is a space-efficient probabilistic data structure used to test whether an element is a member of a set. It can provide quick membership checks with a small probability of false positives, but it does not produce false negatives.

In this exercise, you will implement a simple Bloom Filter with the following functionality:
- Add elements to the filter.
- Check if an element might be present in the filter.

## Instructions

1. Implement the `NewBloomFilter` function to initialize a Bloom Filter with a given size and list of hash functions.
2. Implement the `Add` method to insert an element into the filter.
3. Implement the `Contains` method to check if an element might be present in the filter.

### Constraints
- The filter should use multiple hash functions.
- The filter may return false positives but should not return false negatives.

### Example Usage

```go
package main

import (
    "bloomfilter"
    "fmt"
    "hash/fnv"
)

func main() {
    hashFunctions := []hash.Hash32{fnv.New32a(), fnv.New32()}
    bf := bloomfilter.NewBloomFilter(100, hashFunctions)

    bf.Add("apple")
    bf.Add("banana")

    fmt.Println(bf.Contains("apple"))    // true
    fmt.Println(bf.Contains("banana"))   // true
    fmt.Println(bf.Contains("grape"))    // false (with a chance of false positive)
}
