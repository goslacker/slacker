package grpcgatewayx

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type CustomerHandler struct {
	Method  string
	Path    string
	Handler runtime.HandlerFunc
}

func (c *CustomerHandler) Key() HandlerKey {
	return HandlerKey(fmt.Sprintf("%s|%s", c.Method, c.Path))
}

type HandlerKey string

func (h HandlerKey) Method() string {
	return string(h)[:strings.Index(string(h), "|")]
}

func (h HandlerKey) Path() string {
	return string(h)[strings.Index(string(h), "|")+1:]
}

func NewPProfHandler(pathPrefix string) CustomerHandler {
	return CustomerHandler{
		Method: "GET",
		Path:   strings.TrimSuffix(pathPrefix, "/") + "/debug/pprof/{path}",
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			switch pathParams["path"] {
			case "":
				pprof.Index(w, r)
			case "cmdline":
				pprof.Cmdline(w, r)
			case "profile":
				pprof.Profile(w, r)
			case "symbol":
				pprof.Symbol(w, r)
			case "trace":
				pprof.Trace(w, r)
			default:
				pprof.Handler(pathParams["path"]).ServeHTTP(w, r)
			}
		},
	}
}
