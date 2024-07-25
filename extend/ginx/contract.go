package ginx

import "net/http"

type Router interface {
	Use(...any) Router

	Handle(string, string, ...any) Router
	Any(string, ...any) Router
	GET(string, ...any) Router
	POST(string, ...any) Router
	DELETE(string, ...any) Router
	PATCH(string, ...any) Router
	PUT(string, ...any) Router
	OPTIONS(string, ...any) Router
	HEAD(string, ...any) Router
	Match([]string, string, ...any) Router
	Group(string, ...any) Router

	StaticFile(string, string) Router
	StaticFileFS(string, string, http.FileSystem) Router
	Static(string, string) Router
	StaticFS(string, http.FileSystem) Router
}
