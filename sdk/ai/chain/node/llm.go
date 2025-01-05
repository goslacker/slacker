package node

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/sdk/ai/chain"
	"github.com/goslacker/slacker/sdk/ai/client"
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
		llm.NodeInfo.ID = id
	}
}

func WithVariables(variables ...chain.Variable) func(*LLM) {
	return func(llm *LLM) {
		llm.NodeInfo.Variables = variables
	}
}

func WithSystemPrompt(SystemPrompt string) func(*LLM) {
	return func(llm *LLM) {
		llm.PromptTpl = SystemPrompt
	}
}

func WithPromptTplKey(systemPromptTplKey string) func(*LLM) {
	return func(llm *LLM) {
		llm.PromptTplKey = systemPromptTplKey
	}
}

func NewLLM(name string, model string, inputKey string, opts ...func(*LLM)) *LLM {
	l := &LLM{
		NodeInfo: chain.NodeInfo{
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
		Model:    model,
		InputKey: inputKey,
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
	chain.NodeInfo
	Model         string          `json:"model"`
	EnableHistory bool            `json:"enableHistory"`
	Limit         int             `json:"limit"`
	PromptTpl     string          `json:"promptTpl"`
	Temperature   *float32        `json:"temperature"`
	InputKey      string          `json:"inputKey"`
	Client        client.AIClient `json:"-"`
	Tools         []LLMTool       `json:"tools"`
	NextID        string          `json:"nextId"`
	PromptTplKey  string          `json:"promptTplKey"`
	NoSystem      bool            `json:"noSystem"`
}

func (l *LLM) prepareHisotryWithSystem(ctx chain.Context, history []client.Message) (newHistory []client.Message, err error) {
	if len(history) == 0 && l.PromptTpl != "" {
		var prompt string
		prompt, err = renderPrompt(l.PromptTpl, ctx.GetAllParams(), nil)
		if err != nil {
			return
		}
		history = append(history, client.Message{
			Role:    "system",
			Content: prompt,
		})
	}

	return history, nil
}

func (l *LLM) Run(ctx chain.Context) (nextID string, err error) {
	if l.PromptTpl == "" && l.PromptTplKey == "" {
		return "", errors.New("prompt tpl is required")
	}

	if l.PromptTplKey != "" {
		l.PromptTpl = ctx.GetParam(l.PromptTplKey).(string)
	}

	if l.InputKey == "" {
		return "", errors.New("input key is required")
	}
	if ctx.GetParam(l.InputKey) == nil {
		return "", errors.New("input is required")
	}

	var history []client.Message
	var setHistory func(messages ...client.Message)
	if l.EnableHistory {
		history = ctx.GetHistory(l.Limit)
		setHistory = ctx.SetHistory
	}

	if !l.NoSystem {
		history, err = l.prepareHisotryWithSystem(ctx, history)
		if err != nil {
			return
		}
	}

	history = append(history, client.Message{
		Role:    "user",
		Content: ctx.GetParam(l.InputKey).(string),
	})

	req := &client.ChatCompletionReq{
		Model:       l.Model,
		Temperature: l.Temperature,
		Tools: slicex.MustMap(l.Tools, func(tool LLMTool) client.Tool {
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

	if l.NoSystem {
		var prompt string
		prompt, err = renderPrompt(l.PromptTpl, ctx.GetAllParams(), history)
		if err != nil {
			return
		}
		req.Messages = []client.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		}
	} else {
		req.Messages = history
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

		req = &client.ChatCompletionReq{
			Model:       l.Model,
			Temperature: l.Temperature,
			Tools: slicex.MustMap(l.Tools, func(tool LLMTool) client.Tool {
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

		if l.NoSystem {
			var prompt string
			prompt, err = renderPrompt(l.PromptTpl, ctx.GetAllParams(), history)
			if err != nil {
				return
			}
			req.Messages = []client.Message{
				{
					Role:    "user",
					Content: prompt,
				},
			}
		} else {
			req.Messages = history
		}

		resp, err = l.Client.ChatCompletion(req)

		if err != nil {
			return
		}
	}

	if l.EnableHistory {
		setHistory(append(history, resp.Choices[0].Message)...)
	}

	ctx.SetParam(fmt.Sprintf("%s.%s", l.NodeInfo.GetID(), l.NodeInfo.VariableNames()[0]), resp.Choices[0].Message.Content)
	nextID = l.NextID
	return
}
