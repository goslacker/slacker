package annotator

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"net/http"
)

const (
	ClaimsKey = "x-claims"
)

func PassAuthResult(_ context.Context, req *http.Request) metadata.MD {
	c := req.Context().Value("claims")
	if c != nil {
		b, err := json.Marshal(c)
		if err != nil {
			slog.Warn("failed to marshal claims", "error", err)
			return nil
		}
		return metadata.Pairs(ClaimsKey, string(b))
	}
	return nil
}
