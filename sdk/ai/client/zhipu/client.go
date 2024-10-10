package zhipu

import (
	"errors"
	"fmt"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/core/mapx"
	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/core/tool"
	"github.com/goslacker/slacker/sdk/ai/client"
	"log/slog"
	"net/http"
)

func init() {
	client.Register("glm-4-0520", NewClient)
	client.Register("glm-4v", NewClient)
}

func NewClient(apiKey string) client.AIClient {
	return &Client{
		apiKey: apiKey,
		httpClient: httpClient.NewClient(
			httpClient.WithBaseUrl("https://open.bigmodel.cn/api/paas/v4"),
			httpClient.WithHeader(http.Header{
				"Authorization": {"Bearer " + apiKey},
				"Content-Type":  {"application/json"},
			}),
		),
	}
}

type Client struct {
	apiKey     string
	httpClient *httpClient.Client
}

func (c *Client) ChatCompletion(req *client.ChatCompletionReq) (resp *client.ChatCompletionResp, err error) {
	request := ChatCompletionReq{
		Model: req.Model,
		Messages: slicex.Map(req.Messages, func(item client.Message) Message {
			m := Message{
				Role:       item.Role,
				ToolCallID: item.ToolCallID,
			}
			switch x := item.Content.(type) {
			case string:
				m.Content = x
			case []client.Content:
				m.Content = slicex.Map(x, func(item client.Content) Content {
					c := Content{
						Text: item.Text,
						Type: item.Type,
					}
					if item.Type == "image_url" {
						c.ImageUrl = &ImageUrl{
							Url: item.ImageUrl,
						}
					}
					return c
				})
			case []*client.Content:
				m.Content = slicex.Map(x, func(item *client.Content) Content {
					c := Content{
						Text: item.Text,
						Type: item.Type,
					}
					if item.Type == "image_url" {
						c.ImageUrl = &ImageUrl{
							Url: item.ImageUrl,
						}
					}
					return c
				})
			default:
				panic(errors.New("unsupported type"))
			}
			return m
		}),
		MaxTokens: req.MaxTokens,
		Stop:      req.Stop,
		Tools: slicex.Map(req.Tools, func(item client.Tool) Tool {
			t := Tool{
				Type: string(item.Type),
			}
			switch item.Type {
			case client.ToolTypeFunction:
				t.Function = &Function{
					Name:        item.Function.Name,
					Description: item.Function.Description,
					Parameters: Parameters{
						Type: item.Function.Parameters.Type,
						Properties: mapx.Map(item.Function.Parameters.Properties, func(key string, value client.Property) (newKey string, newValue Property) {
							return key, Property{
								Description: value.Description,
								Type:        value.Type,
								Enum:        value.Enum,
							}
						}),
						Required: item.Function.Parameters.Required,
					},
				}
			case client.ToolTypeRetrieval:
				t.Retrieval = &Retrieval{
					KnowledgeID:    item.Retrieval.KnowledgeID,
					PromptTemplate: item.Retrieval.PromptTemplate,
				}
			case client.ToolTypeWebSearch:
				t.WebSearch = &WebSearch{
					Enable:       item.WebSearch.Enable,
					SearchQuery:  item.WebSearch.SearchQuery,
					SearchResult: item.WebSearch.SearchResult,
				}
			}

			return t
		}),
		UserID: req.User,
	}

	if req.ToolChoice != nil {
		request.ToolChoice = req.ToolChoice.(string)
	}
	if req.Temperature != nil {
		if *req.Temperature <= 0 {
			request.Temperature = tool.Reference(float32(0.1))
		} else if *req.Temperature >= 1 {
			request.Temperature = tool.Reference(float32(0.99))
		} else {
			request.Temperature = req.Temperature
		}
	}

	if req.TopP != nil {
		if *req.TopP <= 0 {
			request.TopP = tool.Reference(float32(0.1))
		} else if *req.TopP >= 1 {
			request.TopP = tool.Reference(float32(0.99))
		} else {
			request.TopP = req.TopP
		}
	}

	response, err := c.httpClient.PostJson("chat/completions", request)
	if err != nil {
		return
	}
	r := &ChatCompletionResp{}
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		err = fmt.Errorf("request zhipu <chat/completions> failed: %s", string(r))
	} else {
		err = response.ScanJson(r)
	}
	if err != nil {
		return
	}

	if r.Error != nil {
		return nil, fmt.Errorf("request zhipu <chat/completions> failed: code: %s, message: %s", r.Error.Code, r.Error.Message)
	}

	resp = &client.ChatCompletionResp{
		ID: r.ID,
		Choices: slicex.Map(r.Choices, func(item Choice) client.Choice {
			return client.Choice{
				Message: client.Message{
					Role:    item.Message.Role,
					Content: tool.Ternary(item.Message.Content == "", nil, item.Message.Content),
					ToolCalls: slicex.Map(item.Message.ToolCalls, func(item ToolCall) client.ToolCall {
						return client.ToolCall{
							ID:   item.ID,
							Type: client.ToolType(item.Type),
							Function: client.Function{
								Name:      item.Function.Name,
								Arguments: item.Function.Arguments,
							},
						}
					}),
				},
				FinishReason: item.FinishReason,
				Index:        item.Index,
			}
		}),
		Created: r.Created,
		Model:   r.Model,
		Usage: client.Usage{
			CompletionTokens: r.Usage.CompletionTokens,
			PromptTokens:     r.Usage.PromptTokens,
			TotalTokens:      r.Usage.TotalTokens,
		},
		WebSearch: slicex.Map(r.WebSearch, func(item WebSearchResp) client.WebSearchResp {
			return client.WebSearchResp{
				Title:   item.Title,
				Link:    item.Link,
				Content: item.Content,
				Icon:    item.Icon,
				Media:   item.Media,
			}
		}),
	}

	slog.Debug("usage", "completion_tokens", resp.Usage.CompletionTokens, "prompt_tokens", resp.Usage.PromptTokens, "total_tokens", resp.Usage.TotalTokens)

	return
}
