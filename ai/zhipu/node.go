package zhipu

import (
	"fmt"
	"github.com/goslacker/slacker/ai"
	"github.com/goslacker/slacker/ai/chain"
)

func NewZhipuNode(promptTpl string, model string, apiKey string, inputKey string, opts ...func(n *chain.LLMNode)) *ZhipuNode {
	llmNode := chain.NewLLMNode(promptTpl, model, inputKey, opts...)
	z := &ZhipuNode{
		client:  NewClient(apiKey),
		LLMNode: llmNode,
	}

	return z
}

type ZhipuNode struct {
	client *Client
	*chain.LLMNode
}

func (z *ZhipuNode) Run(ctx chain.Context) (nextID string, err error) {
	return z.LLMNode.Run(ctx, func(prompt string) (message ai.Message, err error) {
		req := &ChatCompletionReq{
			Model: z.LLMNode.Model,
			Messages: []Message{
				{
					Role:    "user",
					Content: prompt,
				},
			},
		}
		if z.LLMNode.Temperature != 0 {
			req.Temperature = &z.LLMNode.Temperature
		}
		resp, err := z.client.ChatCompletion(req)
		if err != nil {
			err = fmt.Errorf("request chat completion failed: %w", err)
			return
		}

		m, err := ToStandardMessages(*resp.Choices[0].Message)
		if err != nil {
			return
		}
		return m[0], nil
	})
}
