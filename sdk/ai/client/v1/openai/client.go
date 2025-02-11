package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/goslacker/slacker/core/errx"
	httpClient "github.com/goslacker/slacker/core/httpx/client"
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

func (c *Client) ChatCompletion(req *ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *ChatCompletionResp, err error) {
	return c.ChatCompletionWithCtx(context.Background(), req, opts...)
}

func (c *Client) ChatCompletionWithCtx(ctx context.Context, req *ChatCompletionReq, opts ...func(*client.ReqOptions)) (resp *ChatCompletionResp, err error) {

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

	response, err := httpClient.PostJsonCtx(ctx, c.buildUri("chat/completions"), req)
	if err != nil {
		return
	}

	resp = &ChatCompletionResp{}
	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request openai <chat/completions> failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request openai <chat/completions> failed: %s", string(r)))
		}
	} else {
		err = errx.Wrap(response.ScanJson(resp))
	}
	if err != nil {
		return
	}

	slog.Debug("usage openai", "completion_tokens", resp.Usage.CompletionTokens, "prompt_tokens", resp.Usage.PromptTokens, "total_tokens", resp.Usage.TotalTokens)

	return
}

func (c *Client) CreateBatch(ctx context.Context, items []*BatchReqItem, metadata map[string]any) (resp *BatchResponse, err error) {
	var arr [][]byte
	for _, item := range items {
		tmp, _ := json.Marshal(item)
		arr = append(arr, tmp)
	}
	file := &httpClient.File{
		Name:    "batch.jsonl",
		Content: bytes.Join(arr, []byte("\n")),
	}

	fileId, err := c.Upload(ctx, "batch", file)
	if err != nil {
		return
	}

	data := map[string]any{
		"input_file_id":     fileId,
		"endpoint":          "/v1/chat/completions",
		"completion_window": "24h",
	}
	if metadata != nil {
		data["metadata"] = metadata
	}
	response, err := c.httpClient.PostJsonCtx(ctx, "batches", data)
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

	err = response.ScanJson(&resp)
	if err != nil {
		return
	}

	return
}

func (c *Client) RetrieveBatch(ctx context.Context, batchID string) (resp *BatchResponse, err error) {
	response, err := c.httpClient.PostJsonCtx(ctx, "batches", map[string]any{
		"batch_id": batchID,
	})
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

	err = response.ScanJson(&resp)
	if err != nil {
		return
	}

	return
}

func (c *Client) Upload(ctx context.Context, purpose string, file *httpClient.File) (resp UploadResp, err error) {
	response, err := c.httpClient.PostFormCtx(
		ctx,
		"files",
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

	if response.StatusCode > 400 {
		r, _ := response.GetBody()
		switch response.StatusCode {
		case 429:
			err = errx.Wrap(client.ErrRateLimit, errx.WithMsg(fmt.Sprintf("request openai <files> failed: %s", string(r))))
		default:
			err = errx.Wrap(fmt.Errorf("request openai <files> failed: %s", string(r)))
		}
	}

	err = response.ScanJson(&resp)
	if err != nil {
		return
	}

	return
}

func (c *Client) RetrieveFileContent(ctx context.Context, fileID string) (content io.Reader, err error) {
	response, err := c.httpClient.GetCtx(ctx, "files/"+fileID)
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

	buf := new(bytes.Buffer)
	_, err = response.CopyTo(buf)
	if err != nil {
		return
	}
	content = buf
	return
}

func (c *Client) DeleteFile(ctx context.Context, fileID string) (err error) {
	response, err := c.httpClient.DeleteCtx(ctx, "files/"+fileID)
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
