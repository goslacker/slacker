package httpx

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSSEItem(t *testing.T) {
	raw := "id: 0\nevent: Message\ndata: {\"content\":\"这是一\\n个消息\",\"content_type\":\"text\",\"node_is_finish\":true,\"node_seq_id\":\"0\",\"node_title\":\"消息\"}"
	println(len([]byte(raw)))
	item, err := NewSSEItem(raw)
	require.NoError(t, err)
	fmt.Printf("%+v\n", item)
	println(string(item.Data))
}
