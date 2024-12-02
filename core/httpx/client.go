package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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
	debug bool
}

func (c *Client) wrap(f func() (*http.Response, error)) (resp *Response, err error) {
	res, err := f()
	if err != nil {
		return
	}
	if c.debug {
		spyResponse(res)
	}
	resp = &Response{res}

	return
}

func (c *Client) Debug() *Client {
	return &Client{
		Client: c.Client,
		debug:  true,
	}
}

func (c *Client) Get(url string) (resp *Response, err error) {
	req, err := NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.wrap(func() (*http.Response, error) { return c.Client.Do(req) })
}

func (c *Client) Do(req *http.Request) (*Response, error) {
	return c.wrap(func() (*http.Response, error) { return c.do(req) })
}

func (c *Client) do(req *http.Request) (resp *http.Response, err error) {
	if c.debug {
		spyRequest(req)
	}
	return c.Client.Do(req)
}

func (c *Client) Post(url, contentType string, body io.Reader) (resp *Response, err error) {
	req, err := NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.wrap(func() (*http.Response, error) { return c.do(req) })
}

func (c *Client) PostForm(url string, data url.Values) (resp *Response, err error) {
	req, err := NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.wrap(func() (*http.Response, error) { return c.do(req) })
}

func (c *Client) Head(url string) (resp *Response, err error) {
	req, err := NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.wrap(func() (*http.Response, error) { return c.do(req) })
}

func spyResponse(resp *http.Response) {
	spy := &spyBody{
		ReadCloser: resp.Body,
		respMap: map[string]any{
			"status": resp.Status,
			"header": resp.Header,
			"body":   []byte{},
		},
		debug: true,
	}
	if reqSpy, ok := resp.Request.Body.(*spyBody); ok {
		spy.reqMap = reqSpy.reqMap
	}
	resp.Body = spy
}

func spyRequest(req *http.Request) {
	spy := &spyBody{
		ReadCloser:    req.Body,
		isRequestBody: true,
		reqMap: map[string]any{
			"method": req.Method,
			"url":    req.URL.String(),
			"header": req.Header,
			"body":   []byte{},
		},
	}
	req.Body = spy
}

type spyBody struct {
	io.ReadCloser
	reqMap        map[string]any
	respMap       map[string]any
	isRequestBody bool
	debug         bool
}

func (h *spyBody) Read(p []byte) (n int, err error) {
	n, err = h.ReadCloser.Read(p)
	if n > 0 {
		if h.isRequestBody {
			h.reqMap["body"] = append(h.reqMap["body"].([]byte), p[:n]...)
		} else {
			h.respMap["body"] = append(h.respMap["body"].([]byte), p[:n]...)
		}
	}
	return
}

func (h *spyBody) Close() (err error) {
	if h.debug {
		{
			var b map[string]any
			e := json.Unmarshal(h.reqMap["body"].([]byte), &b)
			if e != nil {
				h.reqMap["body"] = string(h.reqMap["body"].([]byte))
			} else {
				h.reqMap["body"] = b
			}
		}
		{
			var b map[string]any
			e := json.Unmarshal(h.respMap["body"].([]byte), &b)
			if e != nil {
				h.respMap["body"] = string(h.respMap["body"].([]byte))
			} else {
				h.respMap["body"] = b
			}
		}

		slog.Debug(
			"request log",
			"req",
			h.reqMap,
			"resp",
			h.respMap,
		)
	}

	return h.ReadCloser.Close()
}

func (h *spyBody) Write(p []byte) (n int, err error) {
	writer, ok := h.ReadCloser.(io.Writer)
	if !ok {
		return 0, errors.New("not a writer")
	}
	return writer.Write(p)
}
