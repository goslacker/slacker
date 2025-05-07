package slogx

import (
	"context"
	"io"
	"log/slog"
	"time"
)

func NewDelayJsonHandler(w io.Writer, opts *slog.HandlerOptions) *DelayJsonHandler {
	return &DelayJsonHandler{
		JSONHandler: slog.NewJSONHandler(w, opts),
	}
}

// DelayJsonHandler 在指定时间内延迟打印日志
type DelayJsonHandler struct {
	*slog.JSONHandler
	ticker *time.Ticker
}

func (h *DelayJsonHandler) SetInterval(t time.Duration) {
	if h.ticker != nil {
		h.ticker.Reset(t)
	} else {
		h.ticker = time.NewTicker(t)
	}
}

func (h *DelayJsonHandler) Handle(ctx context.Context, r slog.Record) error {
	select {
	case <-h.ticker.C:
		return h.JSONHandler.Handle(ctx, r)
	default:
	}
	return nil
}
