package slicex

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEqual(t *testing.T) {
	require.True(t, SameItem([]int{1, 1, 1}...))
	require.False(t, SameItem([]int{1, 1, 2}...))
}
