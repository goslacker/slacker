package zhipu

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_ChatCompletion(t *testing.T) {
	client := NewClient("")
	resp, err := client.ChatCompletion(&ChatCompletionReq{
		Model: "chatglm_pro",
		Messages: []Message{
			{
				Role:    "user",
				Content: "你好",
			},
		},
	})
	require.NoError(t, err)
	fmt.Printf("%+v\n", resp)
}
