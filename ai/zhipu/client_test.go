package zhipu

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"unsafe"
)

func TestClient_ChatCompletion(t *testing.T) {
	client := NewClient("43fdee09f289c308c45390f7e7963a7a.YAg2csgsVqYAznyS")
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

func TestNormal(t *testing.T) {
	println(unsafe.Sizeof(ZhipuNode{}))
}
