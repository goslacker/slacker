package openai

import (
	"context"
	"fmt"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/sdk/ai/client"
	"log/slog"
	"net/http"
)

func init() {
	client.Register("gpt-4o-2024-11-20", NewClient)
}

func NewClient(apiKey string, options ...func(*client.NewOptions)) client.AIClient {
	opts := &client.NewOptions{}
	for _, o := range options {
		o(opts)
	}

	httpOptions := make([]func(*httpClient.Client), 0, 3)
	if opts.Transport != nil {
		httpOptions = append(httpOptions, httpClient.WithTransport(opts.Transport))
	}
	if opts.BaseUrl != "" {
		httpOptions = append(httpOptions, httpClient.WithBaseUrl(opts.BaseUrl))
	} else {
		httpOptions = append(httpOptions, httpClient.WithBaseUrl("https://api.openai.com/v1"))
	}

	if len(opts.Header) == 0 {
		opts.Header = http.Header{}
	}
	opts.Header.Add("Authorization", "Bearer "+apiKey)

	httpOptions = append(httpOptions, httpClient.WithHeader(opts.Header))

	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient.NewClient(httpOptions...),
	}
}

type Client struct {
	apiKey     string
	httpClient *httpClient.Client
}

func (c *Client) ChatCompletion(req *client.ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *client.ChatCompletionResp, err error) {
	return c.ChatCompletionWithCtx(context.Background(), req, opts...)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *client.ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *client.ChatCompletionResp, err error) {

	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	httpClient := c.httpClient
	if len(opt.Header) > 0 {
		for k, v := range opt.Header {
			httpClient = httpClient.AddHeader(k, v...)
		}
	}

	request := FromStdChatCompletionReq(req)
	response, err := httpClient.PostJsonWithCtx(ctx, "chat/completions", request)
	if err != nil {
		return
	}

	r := &ChatCompletionResp{}
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		err = fmt.Errorf("request openai <chat/completions> failed: %s", string(r))
	} else {
		err = response.ScanJson(r)
	}
	if err != nil {
		return
	}

	if r.Error != nil {
		return nil, fmt.Errorf("request openai <chat/completions> failed: code: %s, message: %s", r.Error.Code, r.Error.Message)
	}

	resp = r.IntoStdChatCompletionResp()

	slog.Debug("usage openai", "completion_tokens", resp.Usage.CompletionTokens, "prompt_tokens", resp.Usage.PromptTokens, "total_tokens", resp.Usage.TotalTokens)

	return
}
