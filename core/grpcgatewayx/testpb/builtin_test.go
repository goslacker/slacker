package testpb

import (
	"net/url"
	"testing"

	"github.com/goslacker/slacker/core/grpcgatewayx"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/stretchr/testify/require"
)

//go:generate protoc -I ./testpb --go_out=./testpb --go_opt=paths=source_relative --go-grpc_out=./testpb --go-grpc_opt=paths=source_relative test.proto

func TestQueryParser(t *testing.T) {
	var msg TestMessage
	var p grpcgatewayx.QueryParser
	err := p.Parse(&msg, url.Values{
		"uint64array": []string{"1", "2", "3"},
		"int32array":  []string{"1", "2", "3"},
	}, &utilities.DoubleArray{
		Encoding: map[string]int{},
		Base:     []int(nil),
		Check:    []int(nil),
	})
	require.NoError(t, err)
	require.Equal(t, []uint64{1, 2, 3}, msg.Uint64Array)
	require.Equal(t, []int32{1, 2, 3}, msg.Int32Array)
}
