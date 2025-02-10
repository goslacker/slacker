package claude

import (
	"context"
	"fmt"
	"github.com/goslacker/slacker/core/errx"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/sdk/ai/client"
	"log/slog"
	"net/http"
)

func init() {
	client.Register("claude-3-5-sonnet-20241022", NewClient)
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
		httpOptions = append(httpOptions, httpClient.WithBaseUrl("https://api.anthropic.com/v1"))
	}

	if len(opts.Header) == 0 {
		opts.Header = http.Header{}
	}
	opts.Header.Add("x-api-key", apiKey)
	opts.Header.Add("anthropic-version", "2023-06-01")

	httpOptions = append(httpOptions, httpClient.WithHeader(opts.Header))

	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient.New(httpOptions...),
	}
}

type Client struct {
	apiKey     string
	httpClient httpClient.Client
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

	response, err := httpClient.PostJsonCtx(ctx, "messages", request)
	if err != nil {
		return
	}
	r := &MessageResp{}
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request claude <messages> failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request claude <messages> failed: %s", string(r)))
		}
	} else {
		err = errx.Wrap(response.ScanJson(r))
	}
	if err != nil {
		return
	}

	resp = r.IntoStdChatCompletionResp()

	slog.Debug("usage claude", "completion_tokens", resp.Usage.CompletionTokens, "prompt_tokens", resp.Usage.PromptTokens, "total_tokens", resp.Usage.TotalTokens)

	return
}
