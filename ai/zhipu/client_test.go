package zhipu

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_ChatCompletion(t *testing.T) {
	client := NewClient("a68b1720a601bf60b0f1c45f38725874.PQDVOfBgtUHyTAI4")
	resp, err := client.ChatCompletion(&ChatCompletionReq{
		Model: "glm-4-0520",
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
