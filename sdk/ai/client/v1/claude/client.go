package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/goslacker/slacker/core/errx"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/sdk/ai/client/v1"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func New(apiKey string, options ...func(*client.NewOptions)) *Client {
	opts := &client.NewOptions{}
	for _, o := range options {
		o(opts)
	}

	httpOptions := make([]func(*httpClient.Client), 0, 3)
	if opts.Transport != nil {
		httpOptions = append(httpOptions, httpClient.WithTransport(opts.Transport))
	}
	if opts.EndPoint != "" {
		httpOptions = append(httpOptions, httpClient.WithBaseUrl(opts.EndPoint))
	} else {
		httpOptions = append(httpOptions, httpClient.WithBaseUrl("https://api.anthropic.com"))
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
		apiVersion: "v1",
	}
}

type Client struct {
	apiKey     string
	httpClient httpClient.Client
	apiVersion string
}

func (c *Client) buildUri(uri string) string {
	return fmt.Sprintf("%s/%s", strings.Trim(c.apiVersion, "/"), strings.Trim(uri, "/"))
}

func (c *Client) prepareHttpClient(opt *client.ReqOptions) httpClient.Client {
	httpClient := c.httpClient.Clone()
	if len(opt.Header) > 0 {
		for k, v := range opt.Header {
			httpClient = httpClient.AddHeader(k, v...)
		}
	}
	return httpClient
}

func (c *Client) scanResp(response *httpClient.Response, resp any) (err error) {
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request claude <messages/batches> failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request claude <messages/batches> failed: %s", string(r)))
		}
	} else {
		err = errx.Wrap(response.ScanJson(resp))
	}
	if err != nil {
		return
	}
	return
}

func (c *Client) RetrieveBatch(ctx context.Context, batchID string, opts ...func(*client.ReqOptions)) (resp *RetrieveBatchResp, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}
	response, err := c.prepareHttpClient(opt).GetCtx(ctx, c.buildUri("messages/batches/"+batchID))
	if err != nil {
		return
	}
	resp = &RetrieveBatchResp{}
	err = c.scanResp(response, resp)

	return
}

func (c *Client) RetrieveBatchResults(ctx context.Context, batchID string, opts ...func(*client.ReqOptions)) (resp []BatchResult, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}
	response, err := c.prepareHttpClient(opt).GetCtx(ctx, c.buildUri("messages/batches/"+batchID+"/results"))
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request claude <messages/batches> failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request claude <messages/batches> failed: %s", string(r)))
		}
		return
	}
	tmp, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	resp, err = slicex.Map(bytes.Split(tmp, []byte("\n")), func(item []byte) (result BatchResult, err error) {
		err = json.Unmarshal(item, &result)
		return
	})

	return
}

func (c *Client) CreateBatch(ctx context.Context, req CreateBatchReq, opts ...func(*client.ReqOptions)) (resp *CreateBatchResp, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	response, err := c.prepareHttpClient(opt).PostJsonCtx(ctx, c.buildUri("messages/batches"), req)
	if err != nil {
		return
	}
	resp = &CreateBatchResp{}
	err = c.scanResp(response, resp)

	return
}

func (c *Client) ChatCompletion(req *MessageReq, opts ...func(*client.ReqOptions)) (resp *MessageResp, err error) {
	return c.ChatCompletionWithCtx(context.Background(), req, opts...)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *MessageReq, opts ...func(*client.ReqOptions)) (resp *MessageResp, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	req.Messages = slicex.MustMapFilter(req.Messages, func(item Message) (Message, bool) {
		if item.Role == "system" {
			if req.System == "" {
				req.System = item.Content
			}
			return item, false
		}
		return item, true
	})

	response, err := c.prepareHttpClient(opt).PostJsonCtx(ctx, c.buildUri("messages"), req)
	if err != nil {
		return
	}
	resp = &MessageResp{}
	err = c.scanResp(response, resp)
	if err != nil {
		return
	}

	slog.Debug("usage claude", "completion_tokens", resp.Usage.OutputTokens, "prompt_tokens", resp.Usage.InputTokens, "total_tokens", resp.Usage.InputTokens+resp.Usage.OutputTokens)

	return
}
