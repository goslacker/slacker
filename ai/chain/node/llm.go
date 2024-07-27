package node

import (
	"errors"
	"fmt"
	"github.com/goslacker/slacker/ai/chain"
	"github.com/goslacker/slacker/ai/client"
	"github.com/goslacker/slacker/extend/slicex"
	"log/slog"
)

func WithEnableHistory() func(*LLM) {
	return func(llm *LLM) {
		llm.EnableHistory = true
	}
}

func WithHistoryLimit(limit int) func(*LLM) {
	return func(llm *LLM) {
		llm.Limit = limit
	}
}

func WithTemperature(temperature float32) func(*LLM) {
	return func(llm *LLM) {
		llm.Temperature = &temperature
	}
}

func WithTools(tools ...LLMTool) func(*LLM) {
	return func(llm *LLM) {
		llm.Tools = tools
	}
}

func WithNextID(id string) func(*LLM) {
	return func(llm *LLM) {
		llm.NextID = id
	}
}

func WithID(id string) func(*LLM) {
	return func(llm *LLM) {
		llm.Info.ID = id
	}
}

func WithVariables(variables ...chain.Variable) func(*LLM) {
	return func(llm *LLM) {
		llm.Info.Variables = variables
	}
}

func NewLLM(name string, model string, promptTpl string, inputKey string, opts ...func(*LLM)) *LLM {
	l := &LLM{
		Info: chain.Info{
			Name: name,
			Type: chain.TypeLLM,
			Variables: []chain.Variable{
				{
					Label: "结果",
					Name:  "result",
					Type:  chain.TypeString,
				},
			},
		},
		Model:     model,
		PromptTpl: promptTpl,
		InputKey:  inputKey,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

type LLMTool struct {
	client.Function
	Func func(params string) string
}

type LLM struct {
	chain.Info
	Model         string
	EnableHistory bool
	Limit         int
	PromptTpl     string
	Temperature   *float32
	InputKey      string
	Client        client.ChatCompletable
	Tools         []LLMTool
	NextID        string
}

func (l *LLM) Run(ctx chain.Context) (nextID string, err error) {
	var history []client.Message
	var setHistory func(messages ...client.Message)
	if l.EnableHistory {
		history = ctx.GetHistory(l.Limit)
		setHistory = ctx.SetHistory
	}

	if l.InputKey == "" {
		return "", errors.New("input key is required")
	}
	if ctx.GetParam(l.InputKey) == nil {
		return "", errors.New("input is required")
	}

	history = append(history, client.Message{
		Role:    "user",
		Content: ctx.GetParam(l.InputKey).(string),
	})
	prompt, err := renderPrompt(l.PromptTpl, ctx.GetAllParams(), history)
	if err != nil {
		return
	}

	req := &client.ChatCompletionReq{
		Model: l.Model,
		Messages: []client.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: l.Temperature,
		Tools: slicex.Map(l.Tools, func(tool LLMTool) client.Tool {
			return client.Tool{
				Type: client.ToolTypeFunction,
				Function: &client.Function{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}),
	}

	resp, err := l.Client.ChatCompletion(req)
	if err != nil {
		return
	}

	for {
		if resp.Choices[0].Message.Content != nil || (resp.Choices[0].Message.ToolCalls == nil || len(resp.Choices[0].Message.ToolCalls) == 0) {
			break
		}

		history = append(history, client.Message{
			Role:      "assistant",
			ToolCalls: resp.Choices[0].Message.ToolCalls,
		})

		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			switch toolCall.Type {
			case client.ToolTypeFunction:
				tool, ok := slicex.Find(l.Tools, func(tool LLMTool) bool {
					return tool.Function.Name == toolCall.Function.Name
				})

				var result string
				if !ok {
					result = "no tool found"
					slog.Warn("function call failed: function not found", "function", toolCall.Function.Name)
				}
				result = tool.Func(toolCall.Function.Arguments)
				slog.Debug("function call", "params", toolCall.Function.Arguments, "result", result)
				history = append(history, client.Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: toolCall.ID,
				})
			}
		}

		prompt, err = renderPrompt(l.PromptTpl, ctx.GetAllParams(), history)
		if err != nil {
			return
		}

		req = &client.ChatCompletionReq{
			Model: l.Model,
			Messages: []client.Message{
				{
					Role:    "user",
					Content: prompt,
				},
			},
			Temperature: l.Temperature,
			Tools: slicex.Map(l.Tools, func(tool LLMTool) client.Tool {
				return client.Tool{
					Type: client.ToolTypeFunction,
					Function: &client.Function{
						Name:        tool.Function.Name,
						Description: tool.Function.Description,
						Parameters:  tool.Function.Parameters,
					},
				}
			}),
		}

		resp, err = l.Client.ChatCompletion(req)

		if err != nil {
			return
		}
	}

	if l.EnableHistory {
		setHistory(append(history, resp.Choices[0].Message)...)
	}

	ctx.SetParam(fmt.Sprintf("%s.%s", l.Info.GetID(), l.Info.VariableNames()[0]), resp.Choices[0].Message.Content)
	nextID = l.NextID
	return
}
