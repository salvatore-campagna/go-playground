package bitset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBitSet(t *testing.T) {
	bs := NewBitSet(100)
	require.NotNil(t, bs)
}

func TestSetAndTest(t *testing.T) {
	bs := NewBitSet(100)

	require.NoError(t, bs.Set(10))
	require.NoError(t, bs.Set(50))
	require.NoError(t, bs.Set(99))

	testCases := []struct {
		pos      int
		expected bool
	}{
		{10, true},
		{50, true},
		{99, true},
		{0, false},
		{20, false},
		{98, false},
	}

	for _, tc := range testCases {
		got, err := bs.Test(tc.pos)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, got)
	}
}

func TestClear(t *testing.T) {
	bs := NewBitSet(100)

	require.NoError(t, bs.Set(10))
	require.NoError(t, bs.Set(50))
	require.NoError(t, bs.Clear(10))

	testCases := []struct {
		pos      int
		expected bool
	}{
		{10, false},
		{50, true},
	}

	for _, tc := range testCases {
		got, err := bs.Test(tc.pos)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, got)
	}
}

func TestCount(t *testing.T) {
	bs := NewBitSet(100)

	require.NoError(t, bs.Set(1))
	require.NoError(t, bs.Set(2))
	require.NoError(t, bs.Set(3))
	require.NoError(t, bs.Set(10))
	require.NoError(t, bs.Set(99))

	assert.Equal(t, 5, bs.Count())
}

func TestEdgeCases(t *testing.T) {
	bs := NewBitSet(64)

	require.NoError(t, bs.Set(0))
	require.NoError(t, bs.Set(63))

	got, err := bs.Test(0)
	require.NoError(t, err)
	assert.True(t, got)

	got, err = bs.Test(63)
	require.NoError(t, err)
	assert.True(t, got)
}

func TestOutOfRange(t *testing.T) {
	bs := NewBitSet(64)

	err := bs.Set(1000)
	assert.Error(t, err)

	err = bs.Clear(1000)
	assert.Error(t, err)

	_, err = bs.Test(1000)
	assert.Error(t, err)
}
