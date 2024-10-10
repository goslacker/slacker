package fmtx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestErrorf(t *testing.T) {
	require.Nil(t, Errorf("test %s %d, %ss %w", "test", 1, "ttt", nil))
	require.NotNil(t, Errorf("test %s %d, %ss %w", "test", 1, "ttt", errors.New("test")))

	require.Nil(t, Errorf("test %s %d, %ss %w %w", "test", 1, "ttt", nil, nil))
	require.NotNil(t, Errorf("test %s %d, %ss %w %w", "test", 1, "ttt", nil, errors.New("test2")))
	require.NotNil(t, Errorf("test %s %d, %ss %w %w", "test", 1, "ttt", errors.New("test1"), errors.New("test2")))

	errors.Join()
	fmt.Printf("%+v\n", Errorf("test %s %d, %ss %w %w", "test", 1, "ttt", errors.New("test1"), errors.New("test2")))
}

func TestXxx(t *testing.T) {
	a := "{{#test#}}{{#test2#}}"
	reg := regexp.MustCompile(`\{\{#.*?#\}\}`)
	newa := reg.ReplaceAllFunc([]byte(a), func(b []byte) []byte {
		b = bytes.TrimPrefix(b, []byte("{{#"))
		b = bytes.TrimSuffix(b, []byte("#}}"))
		b = bytes.TrimSpace(b)
		return []byte(fmt.Sprintf(`{{ index . "%s" }}`, string(b)))
	})
	println(string(newa))
}
