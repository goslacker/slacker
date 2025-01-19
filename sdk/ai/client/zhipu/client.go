package zhipu

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/sdk/ai/client"
)

func init() {
	client.Register("glm-4-0520", NewClient)
	client.Register("glm-4-plus", NewClient)
	client.Register("glm-4v", NewClient)
	client.Register("glm-4v-plus", NewClient)
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
	return c.ChatCompletionWithCtx(context.Background(), req)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *client.ChatCompletionReq) (resp *client.ChatCompletionResp, err error) {
	request := FromStdChatCompletionReq(req)

	response, err := c.httpClient.PostJsonWithCtx(ctx, "chat/completions", request)
	if err != nil {
		return
	}
	r := &GLM4ChatCompletionResp{}
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

	resp = r.IntoStdChatCompletionResp()

	slog.Debug("usage", "completion_tokens", resp.Usage.CompletionTokens, "prompt_tokens", resp.Usage.PromptTokens, "total_tokens", resp.Usage.TotalTokens)

	return
}
