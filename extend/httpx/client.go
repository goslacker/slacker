package httpx

import (
	"io"
	"net/http"
	"net/url"
)

func NewClient(opts ...func(client *http.Client)) *Client {
	c := &http.Client{}
	for _, opt := range opts {
		opt(c)
	}
	return &Client{
		Client: c,
	}
}

type Client struct {
	*http.Client
}

func (c *Client) wrap(f func() (*http.Response, error)) (resp *Response, err error) {
	res, err := f()
	if err != nil {
		return
	}
	resp = &Response{res}
	return
}

func (c *Client) Get(url string) (resp *Response, err error) {
	return c.wrap(func() (*http.Response, error) { return c.Client.Get(url) })
}

func (c *Client) Do(req *http.Request) (*Response, error) {
	return c.wrap(func() (*http.Response, error) { return c.Client.Do(req) })
}

func (c *Client) Post(url, contentType string, body io.Reader) (resp *Response, err error) {
	return c.wrap(func() (*http.Response, error) { return c.Client.Post(url, contentType, body) })
}

func (c *Client) PostForm(url string, data url.Values) (resp *Response, err error) {
	return c.wrap(func() (*http.Response, error) { return c.Client.PostForm(url, data) })
}

func (c *Client) Head(url string) (resp *Response, err error) {
	return c.wrap(func() (*http.Response, error) { return c.Client.Head(url) })
}
