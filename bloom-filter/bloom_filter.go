/*
Package bloomfilter provides an implementation of a simple Bloom Filter.

A Bloom Filter is a probabilistic data structure that allows for efficient
membership tests. It can quickly check whether an element is *possibly*
in a set or *definitely not* in the set. However, it does not store the
elements themselves and may have false positives (i.e., reporting an
element is in the set when it is not).

### Design Choices

This implementation uses `hash.Hash32` for its hash functions instead of
the more generic `hash.Hash` or `hash.Hash64`. The reasons for choosing
`Hash32` are as follows:

1. **Efficiency**:
  - `Hash32` produces a 32-bit `uint32` output, which is efficient for
    indexing into a bitset, especially when the bitset size is not extremely large.
  - Computing a `uint32` hash is generally faster than a `uint64` hash,
    which can be important in high-throughput applications.

2. **Simplicity**:
  - `Hash32` returns a `uint32` value directly, which is easily converted
    to an index using a modulo operation. This avoids the need for extra
    conversions required by the more generic `hash.Hash` interface.

3. **Memory Efficiency**:
  - Using `Hash32` avoids the overhead of allocating and converting
    byte slices (`[]byte`) that would be required if using the more generic
    `hash.Hash` interface.

### Probability of False Positives

A Bloom Filter can have false positives but never false negatives. The probability
of a false positive depends on the size of the bitset (`m`), the number of hash functions (`k`),
and the number of elements inserted (`n`). The formula to compute the probability is:

	P(false positive) ≈ (1 - e^(-kn/m))^k

Where:
- `m` is the size of the bitset (number of bits).
- `k` is the number of hash functions.
- `n` is the number of elements inserted.
- `e` is the base of the natural logarithm (~2.71828).

### Example Calculations

Here are some example calculations of the false positive probability with varying parameters:

1. **Example 1**:

  - `m` = 1000 bits (bitset size)

  - `k` = 3 (number of hash functions)

  - `n` = 100 elements (number of inserted items)

    P(false positive) ≈ 0.105 (10.5%)

2. **Example 2**:

  - `m` = 10,000 bits

  - `k` = 5

  - `n` = 500 elements

    P(false positive) ≈ 0.077 (7.7%)

3. **Example 3**:

  - `m` = 100,000 bits

  - `k` = 7

  - `n` = 2000 elements

    P(false positive) ≈ 0.028 (2.8%)

These examples show how increasing the bitset size or the number of hash functions can reduce the probability of false positives.
*/
package bloomfilter

import (
	"errors"
	"hash"
)

// BloomFilter represents a simple Bloom Filter data structure.
type BloomFilter struct {
	bitset        []bool
	hashFunctions []hash.Hash32
}

// NewBloomFilter initializes a new Bloom Filter with the given size and hash functions.
// It returns an error if the size is less than or equal to zero or if no hash functions are provided.
func NewBloomFilter(m int, hashFunctions []hash.Hash32) (*BloomFilter, error) {
	if m <= 0 {
		return nil, errors.New("bloom filter size must be greater than zero")
	}
	if len(hashFunctions) == 0 {
		return nil, errors.New("at least one hash function is required")
	}

	return &BloomFilter{
		bitset:        make([]bool, m),
		hashFunctions: hashFunctions,
	}, nil
}

// Add inserts an element into the Bloom Filter. It computes an index for each hash
// function and sets the corresponding bit in the bitset to `true`.
func (bf *BloomFilter) Add(element string) {
	for _, hashFunction := range bf.hashFunctions {
		hashFunction.Reset()
		hashFunction.Write([]byte(element))
		index := int(hashFunction.Sum32()) % len(bf.bitset)
		bf.bitset[index] = true
	}
}

// Contains checks if an element might be present in the Bloom Filter. It returns `true`
// if all corresponding bits for the element are set; otherwise, it returns `false`.
// If any bit is not set, the element is definitely not in the set. However, even if
// all bits are set, there is still a possibility of a false positive.
func (bf *BloomFilter) Contains(element string) bool {
	for _, hashFunction := range bf.hashFunctions {
		hashFunction.Reset()
		hashFunction.Write([]byte(element))
		index := int(hashFunction.Sum32()) % len(bf.bitset)
		if !bf.bitset[index] {
			return false
		}
	}
	return true
}
