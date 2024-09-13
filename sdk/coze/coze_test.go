package coze

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestXxx(t *testing.T) {
	data := `{"workflow_id": "7398736121049448488","parameters":{"BOT_USER_INPUT": "测试"}}`
	req, err := http.NewRequest("POST", "https://api.coze.cn/v1/workflow/stream_run", strings.NewReader(data))
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
	client := NewClient("")
	msgCh, errCh := client.WorkflowStream(context.Background(), "7398736121049448488", WithParameters(map[string]any{"BOT_USER_INPUT": "测试"}))
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
