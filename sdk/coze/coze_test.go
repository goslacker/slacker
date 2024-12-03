package coze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	data := map[string]any{
		"workflow_id": "7413652077203947572",
		"parameters": map[string]any{
			"history": map[string]string{
				"role":    "user",
				"content": "test",
			},
		},
	}
	j, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", "https://api.coze.cn/v1/workflow/stream_run", bytes.NewReader(j))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	println(string(b))
}

func TestClient_WorkflowStream(t *testing.T) {
	t.Skip()
	client := NewClient("")
	msgCh, errCh := client.WorkflowStream(context.Background(), "7416260052490010665", WithParameters(map[string]any{"history": []map[string]string{
		{"role": "user", "content": "test"},
	}}))

LOOP:
	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				break LOOP
			}
			println(msg)
		case err, ok := <-errCh:
			if !ok {
				break LOOP
			}
			require.NoError(t, err)
		}
	}
}

func TestClient_Workflow(t *testing.T) {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))

	client := NewClient("")
	msg, err := client.Workflow(context.Background(), "7441193295953231906", WithParameters(map[string]any{"chat": `{"role": "user", "content": "test"}`}))
	require.NoError(t, err)
	fmt.Printf("%+v\n", msg)
	// println(msg)
}
