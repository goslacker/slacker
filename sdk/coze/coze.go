package coze

import (
	"net/http"

	"github.com/goslacker/slacker/core/httpx"
)

const baseUrl httpx.URL = "https://api.coze.cn"

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: httpx.NewClient(),
	}
}

type Client struct {
	token      string
	httpClient *httpx.Client
}

func (c *Client) makeRequest(method string, uri string, data any) *http.Request {
	r, err := httpx.NewRequest(method, baseUrl.Append(uri), data)
	if err != nil {
		panic(err)
	}
	r.Header.Add("Authorization", "Bearer "+c.token)
	return r
}

func (c *Client) Debug() *Client {
	return &Client{
		token:      c.token,
		httpClient: c.httpClient.Debug(),
	}
}
