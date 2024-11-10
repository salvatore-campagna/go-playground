# Hints for Implementing a Bloom Filter

- You may find the `hash` package helpful for creating hash functions.
- Remember that a Bloom Filter uses multiple hash functions to determine which bits to set.
- The `fnv` package provides a non-cryptographic hash function that is fast and suitable for this exercise.
- The size of the filter determines how likely false positives are, so consider this when implementing `NewBloomFilter`.
