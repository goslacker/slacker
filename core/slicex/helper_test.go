package slicex

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEqual(t *testing.T) {
	require.True(t, SameItem([]int{1, 1, 1}...))
	require.False(t, SameItem([]int{1, 1, 2}...))
}

func TestIndex(t *testing.T) {
	a := []byte("fjdlasjfkd你好sjalfdjsal")
	b := []byte("你好")
	require.Equal(t, 10, Index(a, b...))

	a = []byte("fjdlasjfkd好你sjalfdjsal")
	b = []byte("你好")
	require.Equal(t, -1, Index(a, b...))
}
