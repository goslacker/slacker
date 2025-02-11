package client

import (
	"errors"
	"net/http"
)

var (
	ErrRateLimit = errors.New("rate limit exceeded")
)

type NewOptions struct {
	Transport *http.Transport
	Header    http.Header
	EndPoint  string
}

func WithHttpTransport(transport *http.Transport) func(*NewOptions) {
	return func(opts *NewOptions) {
		opts.Transport = transport
	}
}

func WithHttpHeader(header http.Header) func(*NewOptions) {
	return func(opts *NewOptions) {
		opts.Header = header
	}
}

func WithEndpoint(endpoint string) func(*NewOptions) {
	return func(opts *NewOptions) {
		opts.EndPoint = endpoint
	}
}

type ReqOptions struct {
	Header http.Header
}

func WithReqHeader(header http.Header) func(*ReqOptions) {
	return func(opts *ReqOptions) {
		opts.Header = header
	}
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
