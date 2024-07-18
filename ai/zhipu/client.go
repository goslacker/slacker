package zhipu

import (
	"fmt"
	"github.com/goslacker/slacker/extend/httpx/client"
	"net/http"
)

const baseUrl = "https://open.bigmodel.cn/api/paas/v4"

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: client.NewClient(
			client.WithBaseUrl(baseUrl),
			client.WithHeader(http.Header{
				"Authorization": {"bearer " + apiKey},
				"Content-Type":  {"application/json"},
			}),
		),
	}
}

type Client struct {
	apiKey     string
	httpClient *client.Client
}

func (c *Client) ChatCompletion(req *ChatCompletionReq) (resp *ChatCompletionResp, err error) {
	response, err := c.httpClient.PostJson("chat/completions", req)
	if err != nil {
		return
	}
	if response.StatusCode > 400 {
		var r []byte
		r, err = response.GetBody()
		if err != nil {
			return
		}
		err = fmt.Errorf("request <chat/completions> failed: %s", string(r))
	} else {
		resp = &ChatCompletionResp{}
		err = response.ScanJson(resp)
	}

	return
}
