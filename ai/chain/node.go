package chain

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/goslacker/slacker/ai"
)

type NodeInfo struct {
	ID string //uuid
}

func (m *NodeInfo) GetID() string {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return m.ID
}

func WithParamKeys(paramKeys []string) func(n *LLMNode) {
	return func(n *LLMNode) {
		if len(n.ParamKeys) > 0 {
			n.ParamKeys = append(n.ParamKeys, paramKeys...)
		} else {
			n.ParamKeys = paramKeys
		}
	}
}

func WithEnableHistory() func(n *LLMNode) {
	return func(n *LLMNode) {
		n.EnableHistory = true
	}
}

func WithLimit(limit int) func(n *LLMNode) {
	return func(n *LLMNode) {
		n.Limit = limit
	}
}

func WithTemperature(temperature float32) func(n *LLMNode) {
	if temperature <= 0 || temperature > 1 {
		panic(errors.New("temperature can only gt 0 and lt 1"))
	}
	return func(n *LLMNode) {
		n.Temperature = temperature
	}
}

func WithNextID(nextID string) func(n *LLMNode) {
	return func(n *LLMNode) {
		n.NextID = nextID
	}
}

func WithOutputKey(key string) func(n *LLMNode) {
	return func(n *LLMNode) {
		n.OutputKey = key
	}
}

func NewLLMNode(promptTpl string, model string, inputKey string, opts ...func(n *LLMNode)) *LLMNode {
	n := &LLMNode{
		PromptTpl: promptTpl,
		Model:     model,
		InputKey:  inputKey,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

type LLMNode struct {
	NodeInfo
	ParamKeys     []string
	EnableHistory bool
	Limit         int
	InputKey      string
	PromptTpl     string
	Model         string
	Temperature   float32
	OutputKey     string
	NextID        string
}

func (l *LLMNode) Run(ctx Context, process func(prompt string) (message ai.Message, err error)) (nextID string, err error) {
	var history []ai.Message
	var setHistory func(messages ...ai.Message)
	if c, ok := ctx.(ChatContext); !ok && l.EnableHistory {
		err = fmt.Errorf("required chain.Context to manager history")
		return
	} else if l.EnableHistory {
		history = c.GetHistory(l.Limit)
		setHistory = c.SetHistory
	}

	params := ctx.GetParams(l.ParamKeys)

	history = append(history, ai.Message{
		Role:    "user",
		Content: params[l.InputKey].(string),
	})
	prompt, err := ai.RenderPrompt(l.PromptTpl, params, history)

	message, err := process(prompt)
	if err != nil {
		return
	}

	if l.EnableHistory {
		history = append(history, message)
		setHistory(history...)
	}

	ctx.SetParam(fmt.Sprintf("%s.%s", l.GetID(), l.OutputKey), message.Content)
	nextID = l.NextID
	return
}
