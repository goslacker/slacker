package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"slices"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type ResponseLogger struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (r *ResponseLogger) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseLogger) Write(body []byte) (int, error) {
	r.Body = body
	return r.ResponseWriter.Write(body)
}

func GenLogReqAndRespMiddleware(ignores []string) func(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			if slices.Contains(ignores, r.URL.Path) {
				next(w, r, pathParams)
				return
			}
			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			r.Body = io.NopCloser(bytes.NewReader(reqBody))

			w = &ResponseLogger{ResponseWriter: w}

			// 调用下一个处理函数
			next(w, r, pathParams)

			rMap := map[string]any{
				"method": r.Method,
				"url":    r.URL.String(),
				"header": r.Header,
			}
			{
				var b map[string]any
				err = json.Unmarshal(reqBody, &b)
				if err != nil {
					rMap["body"] = "<" + r.Header.Get("Content-Type") + ">"
				} else {
					rMap["body"] = b
				}
			}

			respMap := map[string]any{
				"status": w.(*ResponseLogger).StatusCode,
				"header": w.Header(),
			}
			{
				var b map[string]any
				err = json.Unmarshal(w.(*ResponseLogger).Body, &b)
				if err != nil {
					rMap["body"] = "<" + w.Header().Get("Content-Type") + ">"
				} else {
					respMap["body"] = b
				}
			}

			slog.Debug(
				"request log",
				"req",
				rMap,
				"resp",
				respMap,
			)
		}
	}
}

// LogReqAndRespMiddleware 是一个中间件，用于打印请求和响应的日志
func LogReqAndRespMiddleware(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = io.NopCloser(bytes.NewReader(reqBody))

		w = &ResponseLogger{ResponseWriter: w}

		// 调用下一个处理函数
		next(w, r, pathParams)

		rMap := map[string]any{
			"method": r.Method,
			"url":    r.URL.String(),
			"header": r.Header,
		}
		{
			var b map[string]any
			err = json.Unmarshal(reqBody, &b)
			if err != nil {
				rMap["body"] = "<" + r.Header.Get("Content-Type") + ">"
			} else {
				rMap["body"] = b
			}
		}

		respMap := map[string]any{
			"status": w.(*ResponseLogger).StatusCode,
			"header": w.Header(),
		}
		{
			var b map[string]any
			err = json.Unmarshal(w.(*ResponseLogger).Body, &b)
			if err != nil {
				rMap["body"] = "<" + w.Header().Get("Content-Type") + ">"
			} else {
				respMap["body"] = b
			}
		}

		slog.Debug(
			"request log",
			"req",
			rMap,
			"resp",
			respMap,
		)
	}
}
