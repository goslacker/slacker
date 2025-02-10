package client

import (
	"context"
	"fmt"
	"github.com/goslacker/slacker/core/slicex"
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

func New(opts ...func(*Client)) Client {
	c := Client{
		cookies: make([]*http.Cookie, 0),
		headers: make(http.Header),
		queries: make(url.Values),
	}

	for _, opt := range opts {
		opt(&c)
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

func (c Client) SetBaseUrl(baseUrl string) (client Client) {
	client = c.Clone()
	client.baseUrl = strings.TrimRight(baseUrl, "/")
	return client
}

func (c Client) Debug() (client Client) {
	client = c.Clone()
	client.debug = true
	return
}

func (c Client) AddCookies(cookies ...*http.Cookie) (client Client) {
	client = c.Clone()
	client.cookies = append(client.cookies, cookies...)
	return
}

func (c Client) SetHeaders(header map[string][]string) (client Client) {
	client = c.Clone()
	client.headers = header
	return
}

func (c Client) AddHeader(key string, values ...string) (client Client) {
	client = c.Clone()
	for _, value := range values {
		client.headers.Add(key, value)
	}

	return
}

func (c Client) SetHeader(key string, values ...string) (client Client) {
	client = c.Clone()
	client.headers[key] = values
	for idx, value := range values {
		if idx == 0 {
			client.headers.Set(key, value)
		} else {
			client.headers.Add(key, value)
		}
	}

	return
}

func (c Client) SetTransport(transport *http.Transport) (client Client) {
	client = c.Clone()
	client.transport = transport
	return
}

func (c Client) SetQueries(queries map[string][]string) (client Client) {
	client = c.Clone()
	client.queries = queries
	return
}

func (c Client) AddQuery(key string, values ...string) (client Client) {
	client = c.Clone()
	for _, value := range values {
		client.queries.Add(key, value)
	}

	return
}

func (c Client) SetQuery(key string, values ...string) (client Client) {
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

func (c Client) Clone() Client {
	q, err := url.ParseQuery(c.queries.Encode())
	if err != nil {
		panic(err)
	}

	cookies, err := slicex.Map(c.cookies, func(item *http.Cookie) (*http.Cookie, error) {
		return http.ParseSetCookie(item.String())
	})
	if err != nil {
		panic(err)
	}

	newClient := Client{
		cookies: cookies,
		headers: c.headers.Clone(),
		queries: q,
		baseUrl: c.baseUrl,
		debug:   c.debug,
	}
	if c.transport != nil {
		newClient.transport = c.transport.Clone()
	}
	return newClient
}

func (c Client) makeRequest(ctx context.Context, method string, uri string, body any) (req *Request, err error) {
	u, err := c.buildUrl(uri)
	if err != nil {
		return
	}
	return NewRequest(ctx, method, u, body, c.cookies, c.headers)
}

func (c Client) Do(request *Request) (resp *Response, err error) {
	httpClient := http.Client{
		Transport: c.transport,
	}
	response, err := httpClient.Do(request.Request)
	if err == nil {
		resp = NewResponse(response)
	}

	c.printDebugLog(request, resp)

	return
}

func (c Client) buildUrl(uri string) (u string, err error) {
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

func (c Client) PostJson(uri string, body any) (resp *Response, err error) {
	return c.PostJsonCtx(context.Background(), uri, body)
}

func (c Client) PostJsonCtx(ctx context.Context, uri string, body any) (resp *Response, err error) {
	return c.SetHeader("Content-Type", "application/json").do(ctx, http.MethodPost, uri, body)
}

func (c Client) Post(uri string, body any) (resp *Response, err error) {
	return c.do(context.Background(), http.MethodPost, uri, body)
}

func (c Client) PutJson(uri string, body any) (resp *Response, err error) {
	return c.PutJsonCtx(context.Background(), uri, body)
}

func (c Client) PutJsonCtx(ctx context.Context, uri string, body any) (resp *Response, err error) {
	return c.SetHeader("Content-Type", "application/json").do(ctx, http.MethodPut, uri, body)
}

func (c Client) Put(uri string, body any) (resp *Response, err error) {
	return c.do(context.Background(), http.MethodPut, uri, body)
}

func (c Client) Delete(uri string) (resp *Response, err error) {
	return c.do(context.Background(), http.MethodDelete, uri, nil)
}

func (c Client) Get(uri string) (resp *Response, err error) {
	return c.do(context.Background(), http.MethodGet, uri, nil)
}

func (c Client) do(ctx context.Context, method string, uri string, body any) (resp *Response, err error) {
	client := c.Clone()

	req, err := client.makeRequest(ctx, method, uri, body)
	if err != nil {
		return
	}

	resp, err = client.Do(req)

	return
}

func (c Client) printDebugLog(req *Request, resp *Response) {
	if c.debug {
		if resp == nil {
			slog.Debug("request debug", "request", req.Info(), "response", "null")
		} else {
			slog.Debug("request debug", "request", req.Info(), "response", resp.Info())
		}
	}
}
