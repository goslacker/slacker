package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/goslacker/slacker/core/errx"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/sdk/ai/client/v1"
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
		httpOptions = append(httpOptions, httpClient.WithBaseUrl("https://api.openai.com"))
	}

	if len(opts.Header) == 0 {
		opts.Header = http.Header{}
	}
	opts.Header.Add("Authorization", "Bearer "+apiKey)

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
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request openai failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request openai failed: %s", string(r)))
		}
	} else {
		err = errx.Wrap(response.ScanJson(resp))
	}
	if err != nil {
		return
	}
	return
}

func (c *Client) ChatCompletion(req *ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *ChatCompletionResp, err error) {
	return c.ChatCompletionWithCtx(context.Background(), req, opts...)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *ChatCompletionResp, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	response, err := c.prepareHttpClient(opt).PostJsonCtx(ctx, c.buildUri("chat/completions"), req)
	if err != nil {
		return
	}

	resp = &ChatCompletionResp{}
	err = c.scanResp(response, resp)
	if err != nil {
		return
	}
	if resp.Error != nil {
		err = errx.New(resp.Error.Error())
	}

	slog.Debug("usage openai", "completion_tokens", resp.Usage.CompletionTokens, "prompt_tokens", resp.Usage.PromptTokens, "total_tokens", resp.Usage.TotalTokens)

	return
}

func (c *Client) CreateBatch(ctx context.Context, items []*BatchReqItem, metadata map[string]any, opts ...func(*client.ReqOptions)) (resp *BatchResponse, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	var arr [][]byte
	for _, item := range items {
		tmp, _ := json.Marshal(item)
		arr = append(arr, tmp)
	}
	file := &httpClient.File{
		Name:    "batch.jsonl",
		Content: bytes.Join(arr, []byte("\n")),
	}

	fileResp, err := c.Upload(ctx, "batch", file, opts...)
	if err != nil {
		return
	}

	data := map[string]any{
		"input_file_id":     fileResp.ID,
		"endpoint":          "/v1/chat/completions",
		"completion_window": "24h",
	}
	if metadata != nil {
		data["metadata"] = metadata
	}

	response, err := c.prepareHttpClient(opt).PostJsonCtx(ctx, c.buildUri("batches"), data)
	if err != nil {
		return
	}

	resp = &BatchResponse{}
	err = c.scanResp(response, resp)
	if err != nil {
		return
	}
	if resp.Error != nil {
		err = errx.New(resp.Error.Error())
	}

	return
}

func (c *Client) RetrieveBatch(ctx context.Context, batchID string, opts ...func(*client.ReqOptions)) (resp *BatchResponse, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	response, err := c.prepareHttpClient(opt).GetCtx(ctx, c.buildUri("batches/"+batchID))
	if err != nil {
		return
	}

	resp = &BatchResponse{}
	err = c.scanResp(response, resp)
	if err != nil {
		return
	}
	if resp.Error != nil {
		err = errx.New(resp.Error.Error())
	}

	return
}

func (c *Client) Upload(ctx context.Context, purpose string, file *httpClient.File, opts ...func(*client.ReqOptions)) (resp *UploadResp, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	response, err := c.prepareHttpClient(opt).PostFormCtx(
		ctx,
		c.buildUri("files"),
		map[string]string{
			"purpose": purpose,
		},
		map[string]*httpClient.File{
			"file": file,
		},
	)
	if err != nil {
		return
	}

	resp = &UploadResp{}
	err = c.scanResp(response, resp)
	if err != nil {
		return
	}
	if resp.Error != nil {
		err = errx.New(resp.Error.Error())
	}

	return
}

func (c *Client) RetrieveFileContent(ctx context.Context, fileID string, opts ...func(*client.ReqOptions)) (resp []BatchResult, err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	uri := c.buildUri("files/" + fileID + "/content")
	response, err := c.prepareHttpClient(opt).GetCtx(ctx, uri)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request openai <%s> failed: %s", uri, string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request openai <%s> failed: %s", uri, string(r)))
		}
		return
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	resp, err = slicex.Map(bytes.Split(bytes.Trim(content, "\n"), []byte("\n")), func(item []byte) (result BatchResult, err error) {
		err = json.Unmarshal(item, &result)
		return
	})

	return
}

func (c *Client) DeleteFile(ctx context.Context, fileID string, opts ...func(*client.ReqOptions)) (err error) {
	opt := &client.ReqOptions{}
	for _, o := range opts {
		o(opt)
	}

	response, err := c.prepareHttpClient(opt).DeleteCtx(ctx, c.buildUri("files/"+fileID))
	if err != nil {
		return
	}

	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request openai <files> failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request openai <files> failed: %s", string(r)))
		}
	}

	m := make(map[string]any)
	err = response.ScanJson(&m)
	if err != nil {
		return
	}

	if m["deleted"] == false {
		err = errx.New("delete failed")
	}

	return
}
