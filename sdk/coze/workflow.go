package coze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goslacker/slacker/extend/slicex"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func WithParameters(params map[string]any) func(w *workflowReq) {
	return func(w *workflowReq) {
		w.Parameters = params
	}
}

func WithBotID(botID string) func(w *workflowReq) {
	return func(w *workflowReq) {
		w.BotID = botID
	}
}

func WithExt(ext map[string]string) func(w *workflowReq) {
	return func(w *workflowReq) {
		w.Ext = ext
	}
}

type workflowReq struct {
	WorkflowID string            `json:"workflow_id"`
	Parameters map[string]any    `json:"parameters,omitempty"`
	BotID      string            `json:"bot_id,omitempty"`
	Ext        map[string]string `json:"ext,omitempty"`
}

func (c *Client) WorkflowStream(ctx context.Context, workflowID string, opts ...func(w *workflowReq)) (msg chan string, err chan error) {
	ctx, cancel := context.WithCancel(ctx)
	req := &workflowReq{WorkflowID: workflowID}
	for _, opt := range opts {
		opt(req)
	}
	msg = make(chan string)
	err = make(chan error)
	r := c.makeRequest(http.MethodPost, "/v1/workflow/stream_run", req)
	r = r.WithContext(ctx)

	go func() {
		defer func() {
			close(msg)
			close(err)
			cancel()
		}()

		resp, e := c.httpClient.Do(r)
		if e != nil {
			err <- fmt.Errorf("run workflow stream request failed: %w", e)
			return
		}

		records := make(map[string][]*fragment)
		dataCh, errCh := resp.ScanSSEAsync(ctx)
	LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-errCh:
				if !ok {
					break LOOP
				}
				err <- fmt.Errorf("run workflow stream request failed: %w", e)
				return
			case item, ok := <-dataCh:
				if !ok {
					break LOOP
				}
				f := &fragment{}
				e = item.ScanData(f)
				if e != nil {
					err <- fmt.Errorf("scan data failed: %w <%#v>", e, item)
					return
				}
				records[f.NodeTitle] = append(records[f.NodeTitle], f)
				m, ok := isComplete(records[f.NodeTitle])
				if ok {
					delete(records, f.NodeTitle)
					msg <- m
				}
			}
		}
		for _, item := range records {
			if len(item) == 0 {
				slog.Warn("unexpected empty fragments")
				continue
			}
			iJson, _ := json.Marshal(item)
			slog.Error("网络不稳定导致丢包", "fragments", iJson)
			msg <- fmt.Sprintf("网络不稳定导致丢包, 节点名称: %s", item[0].NodeTitle)
		}
	}()

	return
}

func isComplete(fragments []*fragment) (msg string, ok bool) {
	for _, f := range fragments {
		if f.NodeIsFinish {
			expectCount, _ := strconv.Atoi(f.NodeSeqId)
			if len(fragments)-1 == expectCount {
				msg = strings.Join(slicex.Map(fragments, func(item *fragment) string {
					return item.Content
				}), "")
				ok = true
				return
			}
		}
	}
	return
}

type responseErr struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func (e *responseErr) Error() string {
	return fmt.Sprintf("%s %d", e.ErrorMessage, e.ErrorCode)
}

type fragment struct {
	Content      string `json:"content"`
	ContentType  string `json:"content_type"`
	Cost         string `json:"cost"`
	NodeIsFinish bool   `json:"node_is_finish"`
	NodeSeqId    string `json:"node_seq_id"`
	NodeTitle    string `json:"node_title"`
	Token        int    `json:"token"`
}
