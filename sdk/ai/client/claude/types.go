package claude

import (
	"encoding/json"
	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/sdk/ai/client"
)

func FromStdChatCompletionReq(req *client.ChatCompletionReq) *MessageReq {
	system := ""
	m := &MessageReq{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Messages: slicex.MustMapFilter(req.Messages, func(item client.Message) (client.Message, bool) {
			if item.Role == "system" {
				system = item.Content.(string)
				return client.Message{}, false
			}

			c := client.Message{
				Role: item.Role,
			}

			switch x := item.Content.(type) {
			case string:
				c.Content = x
			case []client.Content:
				c.Content = slicex.MustMap(x, func(item client.Content) Content {
					switch item.Type {
					case client.ContentTypeText:
						return Content{
							Type: "text",
							Text: item.Text,
						}
					case client.ContentTypeImageUrl:
						panic("not yet implemented")
					}
					panic("unsupported item type")
				})
			case client.Content:
				switch x.Type {
				case client.ContentTypeText:
					c.Content = Content{
						Type: "text",
						Text: x.Text,
					}
				case client.ContentTypeImageUrl:
					panic("not yet implemented")
				}
			}

			return c, true
		}),
		StopSequences: req.Stop,
		Stream:        req.Stream,
		Temperature:   req.Temperature,
		Tools: slicex.MustMap(req.Tools, func(item client.Tool) Tool {
			panic("not yet implemented")
		}),
		TopP: req.TopP,
	}

	if req.User != "" {
		m.Metadata = &Metadata{
			UserID: req.User,
		}
	}

	if system != "" {
		m.System = system
	}

	if req.ToolChoice != nil {
		panic("not yet implemented")
	}

	return m
}

type MessageReq struct {
	Model         string           `json:"model"`
	MaxTokens     int              `json:"max_tokens"`
	Messages      []client.Message `json:"messages"`
	Metadata      *Metadata        `json:"metadata,omitempty"`
	StopSequences []string         `json:"stop_sequences,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	System        string           `json:"system,omitempty"`
	Temperature   *float32         `json:"temperature,omitempty"`
	Tools         []Tool           `json:"tools,omitempty"`
	ToolChoice    *ToolChoice      `json:"tool_choice,omitempty"`
	TopK          *float32         `json:"top_k,omitempty"`
	TopP          *float32         `json:"top_p,omitempty"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type MessageResp struct {
	ID         string    `json:"id"`
	Content    []Content `json:"content"`
	Model      string    `json:"model"`
	Role       string    `json:"role"`
	StopReason string    `json:"stop_reason"`
	Type       string    `json:"type"`
	Usage      Usage     `json:"usage"`
	Error      *Error    `json:"error"`
}

func (r *MessageResp) IntoStdChatCompletionResp() *client.ChatCompletionResp {
	return &client.ChatCompletionResp{
		ID: r.ID,
		Choices: slicex.MustMap(r.Content, func(item Content) client.Choice {
			c := client.Choice{
				Message: client.Message{
					Role: r.Role,
				},
				FinishReason: r.StopReason,
			}

			var (
				content   string
				ToolCalls []client.ToolCall
			)
			for _, item := range r.Content {
				switch item.Type {
				case "text":
					content = item.Text
				case "tool_use":
					args, _ := json.Marshal(item.Input)
					ToolCalls = append(ToolCalls, client.ToolCall{
						ID:   item.ToolUseID,
						Type: client.ToolTypeFunction,
						Function: client.Function{
							Arguments: string(args),
						},
					})
				}
			}
			c.Message.Content = content
			c.Message.ToolCalls = ToolCalls

			return c
		}),
		Model: r.Model,
		Usage: client.Usage{
			CompletionTokens: r.Usage.OutputTokens,
			PromptTokens:     r.Usage.InputTokens,
			TotalTokens:      r.Usage.InputTokens + r.Usage.OutputTokens,
		},
	}
}

type Usage struct {
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

type ToolChoice struct {
	Type                   string `json:"type"`
	Name                   string `json:"name,omitempty"`
	DisableParallelToolUse *bool  `json:"disable_parallel_tool_use,omitempty"`
}

type Tool struct {
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	InputSchema     InputSchema   `json:"input_schema"`
	Type            string        `json:"type,omitempty"`
	CacheControl    *CacheControl `json:"cache_control,omitempty"`
	DisplayHeightPX int           `json:"display_height_px,omitempty"`
	DisplayWidthPX  int           `json:"display_width_px,omitempty"`
	DisplayNumber   int           `json:"display_number,omitempty"`
}

type CacheControl struct {
	Type string `json:"type"`
}

type InputSchema struct {
	Type       string                     `json:"type"`
	Properties map[string]client.Property `json:"properties"`
	Required   []string                   `json:"required,omitempty"`
}

type Content struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	Source    *ImageSource   `json:"source,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Metadata struct {
	UserID string `json:"user_id"`
}
