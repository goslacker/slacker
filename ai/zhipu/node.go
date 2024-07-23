package zhipu

import (
	"errors"
	"fmt"
	"github.com/goslacker/slacker/ai"
	"github.com/goslacker/slacker/ai/chain"
)

func WithParamKeys(paramKeys []string) func(n *ZhipuNode) {
	return func(n *ZhipuNode) {
		if len(n.paramKeys) > 0 {
			n.paramKeys = append(n.paramKeys, paramKeys...)
		} else {
			n.paramKeys = paramKeys
		}
	}
}

func WithEnableHistory() func(n *ZhipuNode) {
	return func(n *ZhipuNode) {
		n.enableHistory = true
	}
}

func WithLimit(limit int) func(n *ZhipuNode) {
	return func(n *ZhipuNode) {
		n.limit = limit
	}
}

func WithTemperature(temperature float32) func(n *ZhipuNode) {
	if temperature <= 0 || temperature > 1 {
		panic(errors.New("temperature can only gt 0 and lt 1"))
	}
	return func(n *ZhipuNode) {
		n.temperature = temperature
	}
}

func WithNextID(nextID string) func(n *ZhipuNode) {
	return func(n *ZhipuNode) {
		n.nextID = nextID
	}
}

func NewZhipuNode(promptTpl string, model string, apiKey string, inputKey string, opts ...func(n *ZhipuNode)) *ZhipuNode {
	z := &ZhipuNode{
		client:    NewClient(apiKey),
		promptTpl: promptTpl,
		model:     model,
		inputKey:  inputKey,
		outputKey: "result",
		paramKeys: []string{inputKey},
	}
	for _, opt := range opts {
		opt(z)
	}

	return z
}

type ZhipuNode struct {
	chain.NodeInfo
	client        *Client
	paramKeys     []string
	enableHistory bool
	limit         int
	inputKey      string
	promptTpl     string
	model         string
	temperature   float32
	outputKey     string
	nextID        string
}

func (z *ZhipuNode) Run(ctx chain.Context) (nextID string, err error) {
	var history []ai.Message
	var setHistory func(messages ...ai.Message)
	if c, ok := ctx.(chain.ChatContext); !ok && z.enableHistory {
		err = fmt.Errorf("required chain.Context to manager history")
		return
	} else if z.enableHistory {
		history = c.GetHistory(z.limit)
		setHistory = c.SetHistory
	}

	params := ctx.GetParams(z.paramKeys)
	history = append(history, ai.Message{
		Role:    "user",
		Content: params[z.inputKey].(string),
	})
	prompt, err := ai.RenderPrompt(z.promptTpl, params, history)

	req := &ChatCompletionReq{
		Model: z.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	if z.temperature != 0 {
		req.Temperature = &z.temperature
	}
	resp, err := z.client.ChatCompletion(req)
	if err != nil {
		err = fmt.Errorf("request chat completion failed: %w", err)
		return
	}

	if z.enableHistory {
		var m []ai.Message
		m, err = ToStandardMessages(*resp.Choices[0].Message)
		if err != nil {
			return
		}
		history = append(history, m[0])
		setHistory(history...)
	}

	ctx.SetParam(fmt.Sprintf("%s.%s", z.GetID(), z.outputKey), resp.Choices[0].Message.Content)
	nextID = z.nextID
	return
}
