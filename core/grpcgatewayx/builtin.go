package grpcgatewayx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/goslacker/slacker/core/grpcx"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func DefaultResponseRewriter(ctx context.Context, response proto.Message) (any, error) {
	resp := make(map[string]any)
	if _, ok := response.(*emptypb.Empty); ok {
		resp["data"] = nil
	} else {
		resp["data"] = response
	}
	resp["message"] = ""
	return resp, nil
}

func DefaultErrorHandler(_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	resp := make(map[string]any)
	st, ok := status.FromError(err)
	if ok {
		details := st.Details()
		if len(details) > 0 {
			detail := details[0].(*grpcx.ErrorDetail)
			if detail.Code != 0 {
				resp["code"] = detail.Code
			}
			resp["message"] = detail.Message
		}
	}
	if resp["message"] == nil {
		resp["message"] = "unknown error"
	}
	r, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Errorf("marshal err response failed(response=%+v): %w", resp, err).Error()))
		return
	}
	if c, ok := resp["code"]; ok {
		w.WriteHeader(code2http(c.(int32)))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(r)
}

type QueryParser struct {
	runtime.DefaultQueryParser
}

func (q *QueryParser) Parse(msg proto.Message, values url.Values, filter *utilities.DoubleArray) error {
	// 将 'xxx[]' 的字段名转换为 'xxx', 然后调用默认的解析器, 可以达到将 x[]=1 这样的参数从 string 解析为 数字类型
	for k := range values {
		if strings.HasSuffix(k, "[]") {
			values[strings.TrimSuffix(k, "[]")] = values[k]
			delete(values, k)
		}
	}
	return q.DefaultQueryParser.Parse(msg, values, filter)
}
