package tool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelfIP(t *testing.T) {
	ip, err := SelfIP("www.baidu.com:80")
	require.NoError(t, err)
	println(ip)
}
