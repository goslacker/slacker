package grpcgatewayx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/goslacker/slacker/core/grpcx"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
