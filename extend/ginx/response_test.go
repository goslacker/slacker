package ginx

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"reflect"
	"testing"
)

func TestFromResults(t *testing.T) {
	t.Run("test_empty_results", func(t *testing.T) {
		// 测试空返回
		results := []reflect.Value{}
		response := fromResults(results)
		require.IsType(t, &SuccessJsonResponse{}, response)
		require.Equal(t, http.StatusOK, response.(*SuccessJsonResponse).StatusCode)
	})

	t.Run("test_return_error", func(t *testing.T) {
		// 测试返回 error
		results := []reflect.Value{
			reflect.ValueOf(http.StatusInternalServerError),
			reflect.ValueOf(error(nil)),
		}
		response := fromResults(results)
		require.IsType(t, &SuccessJsonResponse{}, response)
		require.Equal(t, http.StatusInternalServerError, response.(*SuccessJsonResponse).StatusCode)

		results = []reflect.Value{
			reflect.ValueOf(http.StatusInternalServerError),
			reflect.ValueOf(errors.New("test error")),
		}
		response = fromResults(results)
		require.IsType(t, &ErrorJsonResponse{}, response)
		require.Equal(t, http.StatusInternalServerError, response.(*ErrorJsonResponse).StatusCode)
		require.Equal(t, "test error", response.(*ErrorJsonResponse).Message)

		results = []reflect.Value{
			reflect.ValueOf(http.StatusInternalServerError),
			reflect.ValueOf(true),
			reflect.ValueOf(errors.New("test error")),
		}
		response = fromResults(results)
		require.IsType(t, &ErrorJsonResponse{}, response)
		require.Equal(t, http.StatusInternalServerError, response.(*ErrorJsonResponse).StatusCode)
		require.Equal(t, "test error", response.(*ErrorJsonResponse).Message)
		require.Equal(t, true, response.(*ErrorJsonResponse).Abort)
	})

	t.Run("test_return_response", func(t *testing.T) {
		resp := &SuccessJsonResponse{}
		// 测试直接返回 Response 接口
		results := []reflect.Value{
			reflect.ValueOf(resp),
		}
		response := fromResults(results)
		require.IsType(t, &SuccessJsonResponse{}, response)
		require.Same(t, resp, response)
	})

	t.Run("test_return_file", func(t *testing.T) {
		// 测试返回文件
		results := []reflect.Value{
			reflect.ValueOf(&NormalFile{
				name:     "test.png",
				mimeType: "image/png",
				content:  []byte{},
			}),
		}
		response := fromResults(results)
		require.IsType(t, &FileResponse{}, response)
		require.Equal(t, "test.png", response.(*FileResponse).File.Name())
	})

	t.Run("test_return_success", func(t *testing.T) {
		// 测试返回成功 JSON 响应
		results := []reflect.Value{
			reflect.ValueOf(http.StatusOK),
			reflect.ValueOf(map[string]any{"key": "value"}),
		}
		response := fromResults(results)
		require.IsType(t, &SuccessJsonResponse{}, response)
		require.Equal(t, "value", response.(*SuccessJsonResponse).Data.(map[string]any)["key"])
	})

	t.Run("test_return_success_and_meta", func(t *testing.T) {
		// 测试返回成功 JSON 响应
		results := []reflect.Value{
			reflect.ValueOf(http.StatusOK),
			reflect.ValueOf(map[string]any{"key": "value"}),
			reflect.ValueOf(Meta(map[string]any{"key": "value1"})),
		}
		response := fromResults(results)
		require.IsType(t, &SuccessJsonResponse{}, response)
		require.Equal(t, "value", response.(*SuccessJsonResponse).Data.(map[string]any)["key"])
		require.Equal(t, "value1", response.(*SuccessJsonResponse).Meta["key"])
	})
}
