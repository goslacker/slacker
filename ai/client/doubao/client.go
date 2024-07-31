package doubao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/goslacker/slacker/ai/client"
	httpClient "github.com/goslacker/slacker/extend/httpx/client"
	"github.com/goslacker/slacker/extend/slicex"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

func init() {
	client.Register("doubao-pro-32k", NewClient)
}

func NewClient(apiKey string) client.AIClient {
	a := strings.Split(apiKey, "|")
	return &Client{
		model:  a[0],
		apiKey: a[1],
		httpClient: httpClient.NewClient(
			httpClient.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
			httpClient.WithHeader(http.Header{
				"Authorization": {"Bearer " + apiKey},
				"Content-Type":  {"application/json"},
			}),
		),
	}
}

type Client struct {
	model      string
	apiKey     string
	httpClient *httpClient.Client
}

func (c *Client) ChatCompletion(req *client.ChatCompletionReq) (resp *client.ChatCompletionResp, err error) {
	clit := arkruntime.NewClientWithApiKey(c.apiKey)
	ctx := context.Background()

	request := model.ChatCompletionRequest{
		Model: c.model,
		Messages: slicex.Map(req.Messages, func(message client.Message) *model.ChatCompletionMessage {
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
		Tools: slicex.Map(req.Tools, func(tool client.Tool) *model.Tool {
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
		Choices: slicex.Map(response.Choices, func(choice *model.ChatCompletionChoice) client.Choice {
			return client.Choice{
				FinishReason: string(choice.FinishReason),
				Index:        int(choice.Index),
				Message: client.Message{
					Content: *choice.Message.Content.StringValue,
					Role:    choice.Message.Role,
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
