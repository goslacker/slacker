package doubao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/sdk/ai/client"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

func init() {
	client.Register("doubao-pro-32k", NewClient)
}

func NewClient(apiKey string, options ...func(newOptions *client.NewOptions)) client.AIClient {
	opts := &client.NewOptions{}
	for _, o := range options {
		o(opts)
	}

	a := strings.Split(apiKey, "|")
	return &Client{
		model:     a[0],
		apiKey:    a[1],
		transport: opts.Transport,
		baseUrl:   opts.BaseUrl,
	}
}

type Client struct {
	model     string
	apiKey    string
	transport *http.Transport
	baseUrl   string
}

func (c *Client) SetBaseUrl(baseUrl string) {
	panic("implement me")
}

func (c *Client) ChatCompletion(req *client.ChatCompletionReq) (resp *client.ChatCompletionResp, err error) {
	return c.ChatCompletionWithCtx(context.Background(), req)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *client.ChatCompletionReq) (resp *client.ChatCompletionResp, err error) {
	var clit *arkruntime.Client
	//不导出就很迷
	if c.transport != nil && c.baseUrl != "" {
		clit = arkruntime.NewClientWithApiKey(c.apiKey, arkruntime.WithHTTPClient(&http.Client{Transport: c.transport}), arkruntime.WithBaseUrl(c.baseUrl))
	} else if c.transport != nil {
		clit = arkruntime.NewClientWithApiKey(c.apiKey, arkruntime.WithHTTPClient(&http.Client{Transport: c.transport}))
	} else if c.baseUrl != "" {
		clit = arkruntime.NewClientWithApiKey(c.apiKey, arkruntime.WithBaseUrl(c.baseUrl))
	} else {
		clit = arkruntime.NewClientWithApiKey(c.apiKey)
	}

	request := model.ChatCompletionRequest{
		Model: c.model,
		Messages: slicex.MustMap(req.Messages, func(message client.Message) *model.ChatCompletionMessage {
			msg := &model.ChatCompletionMessage{
				ToolCallID: message.ToolCallID,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(message.Content.(string)),
				},
			}
			switch message.Role {
			case string(client.RoleUser):
				msg.Role = model.ChatMessageRoleUser
			case string(client.RoleAssistant):
				msg.Role = model.ChatMessageRoleAssistant
			case string(client.RoleSystem):
				msg.Role = model.ChatMessageRoleSystem
			}
			return msg
		}),
		Tools: slicex.MustMap(req.Tools, func(tool client.Tool) *model.Tool {
			t := &model.Tool{
				Type: model.ToolTypeFunction,
				Function: &model.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
			return t
		}),
	}

	if req.Temperature != nil {
		request.Temperature = *req.Temperature
	}

	if req.TopP != nil {
		request.TopP = *req.TopP
	}

	response, err := clit.CreateChatCompletion(ctx, request)
	if err != nil {
		err = fmt.Errorf("standard chat error: %w", err)
		return
	}

	resp = &client.ChatCompletionResp{
		Choices: slicex.MustMap(response.Choices, func(choice *model.ChatCompletionChoice) client.Choice {
			return client.Choice{
				FinishReason: string(choice.FinishReason),
				Index:        int(choice.Index),
				Message: client.Message{
					Content: *choice.Message.Content.StringValue,
					Role:    choice.Message.Role,
					ToolCalls: slicex.MustMap(choice.Message.ToolCalls, func(toolCall *model.ToolCall) client.ToolCall {
						call := client.ToolCall{
							ID:   toolCall.ID,
							Type: client.ToolType(toolCall.Type),
							Function: client.Function{
								Arguments: toolCall.Function.Arguments,
								Name:      toolCall.Function.Name,
							},
						}
						return call
					}),
				},
			}
		}),
		Usage: client.Usage{
			CompletionTokens: response.Usage.CompletionTokens,
			PromptTokens:     response.Usage.PromptTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}

	return
}
