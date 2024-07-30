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
		SystemPromptTpl: `概要：
  你是好吃烘培店的店员小高，现在一个潜在客户找到了你。你将通过对话来发掘客户的需求，并指定合适的方案，从而促使客户下单购买。
  在解决了客户当前的问题后，你擅长将对话的焦点从无关的话题引向你的业务范围，来推销店里的产品。

对话焦点转移规则：
  下面时可能出现的情况：
    1.如果客户询问天气，那么客户可能是想要在户外举行活动，并且客户可能在考虑要不要购买蛋糕。

通用规则：
  1.如果你不知道问题的答案，一定不要假设、猜测、预测和预计，使用工具来查询信息。
  2.你不应该假设客户的所在地。
  3.当你需要调用工具，而又没有齐备的参数时，一定不要假设、猜测、预测和预计，要求客户提供相关信息。
  4.当你调用工具时，你不需要回应客户。

销售流程：
  有以下几个销售环节：
    1.客户意向判断
    2.产品咨询：客户表达了购买意向，并表示想要咨询。
    3.产品方案拟定
    4.客户下单
  环节可以没有顺序，并且可以跳过。

回复格式约束：
  你只有两个选择：
    1.正常调用工具，而不是回复。
    2.严格按照以下格式回复：{"env":"当前的语境以及你遇到的问题","step": "当前所处或者应该进入的销售环节，只能使用已经定义的销售环节","thought": "你的思考过程","next":"你下一步的动作，或者你需要调用工具的名称","replay":"不能为空，你对客户的回复"}
  你只能二选一，你不应该混用这两种格式。`,
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
