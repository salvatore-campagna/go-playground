package bloomfilter_test

import (
	"bloomfilter"
	"hash"
	"hash/fnv"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBloomFilter_ValidInputs(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(1000, hashFunctions)
	require.NoError(t, err)
	assert.NotNil(t, bf)
}

func TestNewBloomFilter_SizeZero(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(0, hashFunctions)
	require.Error(t, err)
	assert.EqualError(t, err, "bloom filter size must be greater than zero")
	assert.Nil(t, bf)
}

func TestNewBloomFilter_EmptyHashFunctions(t *testing.T) {
	bf, err := bloomfilter.NewBloomFilter(1000, []hash.Hash32{})
	require.Error(t, err)
	assert.EqualError(t, err, "at least one hash function is required")
	assert.Nil(t, bf)
}

func TestNewBloomFilter_NegativeSize(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(-1, hashFunctions)
	require.Error(t, err)
	assert.EqualError(t, err, "bloom filter size must be greater than zero")
	assert.Nil(t, bf)
}

func TestBloomFilter_AddAndContains(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(1000, hashFunctions)
	require.NoError(t, err)
	require.NotNil(t, bf)

	bf.Add("apple")
	bf.Add("banana")
	bf.Add("grape")

	assert.True(t, bf.Contains("apple"))
	assert.True(t, bf.Contains("banana"))
	assert.True(t, bf.Contains("grape"))
}

func TestBloomFilter_ContainsNonExistent(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(1000, hashFunctions)
	require.NoError(t, err)
	require.NotNil(t, bf)

	bf.Add("apple")
	bf.Add("banana")

	assert.False(t, bf.Contains("orange"))
	assert.False(t, bf.Contains("pineapple"))
}

func TestBloomFilter_RandomData(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(5000, hashFunctions)
	require.NoError(t, err)
	require.NotNil(t, bf)

	rand.Seed(time.Now().UnixNano())

	randomData := generateRandomStrings(1000, 10)
	for _, str := range randomData {
		bf.Add(str)
	}

	for _, str := range randomData {
		assert.True(t, bf.Contains(str), "Expected '%s' to be in the Bloom Filter", str)
	}
}

func TestBloomFilter_Empty(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(1000, hashFunctions)
	require.NoError(t, err)
	require.NotNil(t, bf)

	// Check that an empty Bloom Filter returns false for any input
	assert.False(t, bf.Contains("apple"))
	assert.False(t, bf.Contains("banana"))
	assert.False(t, bf.Contains("grape"))
}

func TestBloomFilter_VeryRareData(t *testing.T) {
	hashFunctions := []hash.Hash32{fnv.New32(), fnv.New32a()}
	bf, err := bloomfilter.NewBloomFilter(5000, hashFunctions)
	require.NoError(t, err)
	require.NotNil(t, bf)

	// Add some elements
	bf.Add("apple")
	bf.Add("banana")
	bf.Add("grape")

	// Check for an element that was definitely not added
	// Using a very long and unique string to reduce the chance of a false positive
	assert.False(t, bf.Contains("this_is_a_very_long_string_that_is_unlikely_to_collide"))
}

func generateRandomStrings(count, length int) []string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]string, count)
	for i := range result {
		result[i] = randomString(length, charset)
	}
	return result
}

func randomString(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
