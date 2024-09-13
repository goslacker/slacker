package httpx

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestXxx(t *testing.T) {
	data := `{"workflow_id": "739873612104944848","parameters":{"BOT_USER_INPUT": "测试"}}`
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

func TestResponse_ScanSSEAsync(t *testing.T) {
	data := `{"workflow_id": "7398736121049448488","parameters":{"BOT_USER_INPUT": "测试"}}`
	client := NewClient()
	req, err := NewRequest("POST", "https://api.coze.cn/v1/workflow/stream_run", data)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer ")
	resp, err := client.Do(req)
	require.NoError(t, err)

	sseCh, errCh := resp.ScanSSEAsync(context.Background())

LOOP:
	for {
		select {
		case msg, ok := <-sseCh:
			if !ok {
				break LOOP
			}
			fmt.Printf("%+v\n", string(msg.Data))
		case e, ok := <-errCh:
			if !ok {
				break LOOP
			}
			require.NoError(t, e)
		}
	}
}
