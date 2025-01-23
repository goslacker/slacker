package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

func WithBaseUrl(baseUrl string) func(*Client) {
	return func(client *Client) {
		client.baseUrl = baseUrl
	}
}

func WithHeader(header http.Header) func(*Client) {
	return func(client *Client) {
		client.headers = header
	}
}

func WithTransport(transport *http.Transport) func(*Client) {
	return func(client *Client) {
		client.transport = transport
	}
}

func NewClient(opts ...func(*Client)) *Client {
	c := &Client{
		cookies: make([]*http.Cookie, 0),
		headers: make(http.Header),
		queries: make(url.Values),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type Client struct {
	cookies   []*http.Cookie
	headers   http.Header
	clone     int
	queries   url.Values
	debug     bool
	baseUrl   string
	transport *http.Transport
}

func (c *Client) SetBaseUrl(baseUrl string) {
	c.baseUrl = strings.TrimRight(baseUrl, "/")
	return
}

func (c *Client) Debug() (client *Client) {
	client = c.Clone()
	client.debug = true
	return
}

func (c *Client) AddCookies(cookies ...*http.Cookie) (client *Client) {
	client = c.Clone()
	client.cookies = append(client.cookies, cookies...)
	return
}

func (c *Client) SetHeaders(header map[string][]string) (client *Client) {
	client = c.Clone()
	client.headers = header
	return
}

func (c *Client) AddHeader(key string, values ...string) (client *Client) {
	client = c.Clone()
	for _, value := range values {
		client.headers.Add(key, value)
	}

	return
}

func (c *Client) SetHeader(key string, values ...string) (client *Client) {
	client = c.Clone()
	for idx, value := range values {
		if idx == 0 {
			client.headers.Set(key, value)
		} else {
			client.headers.Add(key, value)
		}
	}

	return
}

func (c *Client) SetTransport(transport *http.Transport) (client *Client) {
	client = c.Clone()
	client.transport = transport
	return
}

func (c *Client) SetQueries(queries map[string][]string) (client *Client) {
	client = c.Clone()
	client.queries = queries
	return
}

func (c *Client) AddQuery(key string, values ...string) (client *Client) {
	client = c.Clone()
	for _, value := range values {
		client.queries.Add(key, value)
	}

	return
}

func (c *Client) SetQuery(key string, values ...string) (client *Client) {
	client = c.Clone()
	for idx, value := range values {
		if idx == 0 {
			client.queries.Set(key, value)
		} else {
			client.queries.Add(key, value)
		}
	}

	return
}

func (c *Client) Clone() *Client {
	switch c.clone {
	case 0:
		q, err := url.ParseQuery(c.queries.Encode())
		if err != nil {
			panic(err)
		}
		client := &Client{
			cookies: c.cookies[:],
			headers: c.headers.Clone(),
			clone:   1,
			queries: q,
			debug:   c.debug,
			baseUrl: c.baseUrl,
		}
		if client.transport != nil {
			client.transport = c.transport.Clone()
		}
		return client
	case 1:
		return c
	default:
		return c
	}
}

func (c *Client) Clean() {
	c.cookies = make([]*http.Cookie, 0)
	c.headers = make(http.Header)
}

func (c *Client) makeRequest(ctx context.Context, method string, url string, body any) (req *Request, err error) {
	return NewRequest(ctx, method, url, body, c.cookies, c.headers)
}

func (c *Client) Do(request *Request) (resp *Response, err error) {
	response, err := http.DefaultClient.Do(request.Request)
	if err != nil {
		return
	}
	resp = NewResponse(response)
	return
}

func (c *Client) buildUrl(uri string) (u string, err error) {
	var tmp *url.URL
	if c.baseUrl != "" {
		tmp, err = url.Parse(fmt.Sprintf("%s/%s", c.baseUrl, strings.TrimLeft(uri, "/")))
	} else {
		tmp, err = url.Parse(uri)
	}
	if err != nil {
		return
	}

	if len(c.queries) > 0 {
		tmp.RawQuery = c.queries.Encode()
	}

	u = tmp.String()
	return
}

func (c *Client) PostJson(uri string, body any) (resp *Response, err error) {
	return c.PostJsonWithCtx(context.Background(), uri, body)
}

func (c *Client) PostJsonWithCtx(ctx context.Context, uri string, body any) (resp *Response, err error) {
	return c.SetHeader("Content-Type", "application/json").PostWithCtx(ctx, uri, body)
}

func (c *Client) Post(uri string, body any) (resp *Response, err error) {
	return c.PostWithCtx(context.Background(), uri, body)
}

func (c *Client) PostWithCtx(ctx context.Context, uri string, body any) (resp *Response, err error) {
	client := c.Clone()
	u, err := client.buildUrl(uri)
	if err != nil {
		return
	}

	req, err := client.makeRequest(ctx, http.MethodPost, u, body)
	if err != nil {
		return
	}

	resp, err = client.Do(req)

	if client.debug {
		if resp == nil {
			slog.Debug("request debug", "request", req.Info(), "response", nil)
		} else {
			slog.Debug("request debug", "request", req.Info(), "response", resp.Info())
		}
	}

	return
}

func (c *Client) PutJson(uri string, body any) (resp *Response, err error) {
	return c.PutJsonWithCtx(context.Background(), uri, body)
}

func (c *Client) PutJsonWithCtx(ctx context.Context, uri string, body any) (resp *Response, err error) {
	return c.SetHeader("Content-Type", "application/json").PutWithCtx(ctx, uri, body)
}

func (c *Client) Put(uri string, body any) (resp *Response, err error) {
	return c.PutWithCtx(context.Background(), uri, body)
}

func (c *Client) PutWithCtx(ctx context.Context, uri string, body any) (resp *Response, err error) {
	client := c.Clone()
	u, err := client.buildUrl(uri)
	if err != nil {
		return
	}

	req, err := client.makeRequest(ctx, http.MethodPut, u, body)
	if err != nil {
		return
	}

	resp, err = client.Do(req)

	if client.debug {
		if resp == nil {
			slog.Debug("request debug", "request", req.Info(), "response", nil)
		} else {
			slog.Debug("request debug", "request", req.Info(), "response", resp.Info())
		}
	}

	return
}

func (c *Client) Delete(uri string) (resp *Response, err error) {
	return c.DeleteWithCtx(context.Background(), uri)
}

func (c *Client) DeleteWithCtx(ctx context.Context, uri string) (resp *Response, err error) {
	client := c.Clone()
	u, err := client.buildUrl(uri)
	if err != nil {
		return
	}

	req, err := client.makeRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return
	}

	resp, err = client.Do(req)

	if client.debug {
		if resp == nil {
			slog.Debug("request debug", "request", req.Info(), "response", nil)
		} else {
			slog.Debug("request debug", "request", req.Info(), "response", resp.Info())
		}
	}

	return
}

func (c *Client) Get(uri string) (resp *Response, err error) {
	return c.GetWithCtx(context.Background(), uri)
}

func (c *Client) GetWithCtx(ctx context.Context, uri string) (resp *Response, err error) {
	client := c.Clone()
	u, err := client.buildUrl(uri)
	if err != nil {
		return
	}

	req, err := client.makeRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return
	}

	resp, err = client.Do(req)

	if client.debug {
		if resp == nil {
			slog.Debug("request debug", "request", req.Info(), "response", "null")
		} else {
			slog.Debug("request debug", "request", req.Info(), "response", resp.Info())
		}
	}

	return
}
