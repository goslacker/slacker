package zhipu

import (
	"context"
	"fmt"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/sdk/ai/client"
	"log/slog"
	"net/http"
)

func init() {
	client.Register("glm-4-0520", NewClient)
	client.Register("glm-4-plus", NewClient)
	client.Register("glm-4v", NewClient)
	client.Register("glm-4v-plus", NewClient)
}

func NewClient(apiKey string, options ...func(newOptions *client.NewOptions)) client.AIClient {
	opt := &client.NewOptions{}
	for _, o := range options {
		o(opt)
	}

	httpOptions := make([]func(*httpClient.Client), 0, 3)
	if opt.Transport != nil {
		httpOptions = append(httpOptions, httpClient.WithTransport(opt.Transport))
	}
	if opt.BaseUrl != "" {
		httpOptions = append(httpOptions, httpClient.WithBaseUrl(opt.BaseUrl))
	} else {
		httpOptions = append(httpOptions, httpClient.WithBaseUrl("https://open.bigmodel.cn/api/paas/v4"))
	}
	if len(opt.Header) > 0 {
		opt.Header.Add("Authorization", "Bearer "+apiKey)
	} else {
		opt.Header = http.Header{
			"Authorization": []string{"Bearer " + apiKey},
		}
	}
	httpOptions = append(httpOptions, httpClient.WithHeader(opt.Header))

	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient.New(httpOptions...),
	}
}

type Client struct {
	apiKey     string
	httpClient httpClient.Client
}

func (c *Client) SetBaseUrl(baseUrl string) {
	c.httpClient.SetBaseUrl(baseUrl)
}

func (c *Client) ChatCompletion(req *client.ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *client.ChatCompletionResp, err error) {
	return c.ChatCompletionWithCtx(context.Background(), req, opts...)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *client.ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *client.ChatCompletionResp, err error) {
	options := &client.ReqOptions{}
	for _, o := range opts {
		o(options)
	}

	httpClient := c.httpClient
	if len(options.Header) > 0 {
		for k, v := range options.Header {
			httpClient = httpClient.AddHeader(k, v...)
		}
	}

	request := FromStdChatCompletionReq(req)

	response, err := httpClient.PostJsonCtx(ctx, "chat/completions", request)
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
