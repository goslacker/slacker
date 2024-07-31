package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/goslacker/slacker/ai/chain"
	"github.com/goslacker/slacker/ai/client"
	_ "github.com/goslacker/slacker/ai/client/zhipu"
	"github.com/stretchr/testify/require"
)

func TestLLM(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	llm := &LLM{
		NodeInfo: chain.NodeInfo{
			ID:   "test",
			Name: "test",
			Type: chain.TypeLLM,
			Variables: []chain.Variable{
				{
					Name:  "result",
					Label: "result",
					Type:  chain.TypeString,
				},
			},
		},
		Model:         "glm-4-0520",
		EnableHistory: true,
		Client:        client.New("glm-4-0520", ""),
		Tools: []LLMTool{
			{
				Func: func(params string) string {
					return "晴,20-30摄氏度"
				},
				Function: client.Function{
					Name:        "get_weather_info",
					Description: "根据城市获取天气信息",
					Parameters: &client.Parameters{
						Type: "object",
						Properties: map[string]client.Property{
							"city": {
								Description: "城市名称",
								Type:        "string",
							},
						},
						Required: []string{"city"},
					},
				},
			},
		},
		InputKey: "input",
		PromptTpl: `下面试聊天记录:
{{#history#}}
`,
		NoSystem: true,
	}
	c := chain.NewContext(context.Background())
	history := chain.NewHistory()
	history.Set([]client.Message{
		{
			Role:    "user",
			Content: "明天天气怎么样",
		},
		{
			Role:    "assistant",
			Content: `{"env":"客户询问天气，可能是在考虑是否在户外举办活动，也可能考虑是否购买蛋糕。","step":"客户意向判断","thought":"需要获取天气信息，以判断是否适合户外活动，进而推断客户的需求。","next":"get_weather_info","replay":"请问您所在的城市是哪里呢？"}`,
		},
	}...)
	c = chain.WithHistory(c, history)
	c.SetParam("input", "成都")
	_, err := llm.Run(c)
	require.NoError(t, err)

	j, err := json.Marshal(history.Get(0))
	require.NoError(t, err)
	println(string(j))

	println(c.GetParam(fmt.Sprintf("%s.%s", llm.GetID(), "result")).(string))
}
