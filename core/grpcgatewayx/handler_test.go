package grpcgatewayx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomerHandler_Key(t *testing.T) {
	t.Run("returns method and path joined by pipe", func(t *testing.T) {
		h := &CustomerHandler{Method: "GET", Path: "/api/v1/users"}
		require.Equal(t, "GET|/api/v1/users", h.Key())
	})

	t.Run("returns correct key with POST method", func(t *testing.T) {
		h := &CustomerHandler{Method: "POST", Path: "/api/v1/users"}
		require.Equal(t, "POST|/api/v1/users", h.Key())
	})

	t.Run("returns correct key with empty method and path", func(t *testing.T) {
		h := &CustomerHandler{Method: "", Path: ""}
		require.Equal(t, "|", h.Key())
	})

	t.Run("returns correct key when path contains pipe", func(t *testing.T) {
		h := &CustomerHandler{Method: "GET", Path: "/api/v1/users|detail"}
		require.Equal(t, "GET|/api/v1/users|detail", h.Key())
	})
}

func TestHandlerKey_Method(t *testing.T) {
	t.Run("extracts method from key", func(t *testing.T) {
		h := HandlerKey("GET|/api/v1/users")
		require.Equal(t, "GET", h.Method())
	})

	t.Run("extracts POST method from key", func(t *testing.T) {
		h := HandlerKey("POST|/api/v1/users")
		require.Equal(t, "POST", h.Method())
	})

	t.Run("extracts empty method when key starts with pipe", func(t *testing.T) {
		h := HandlerKey("|/api/v1/users")
		require.Equal(t, "", h.Method())
	})
}

func TestHandlerKey_Path(t *testing.T) {
	t.Run("extracts path from key", func(t *testing.T) {
		h := HandlerKey("GET|/api/v1/users")
		require.Equal(t, "/api/v1/users", h.Path())
	})

	t.Run("extracts path containing pipe character", func(t *testing.T) {
		h := HandlerKey("GET|/api/v1/users|detail")
		require.Equal(t, "/api/v1/users|detail", h.Path())
	})

	t.Run("extracts empty path when key ends with pipe", func(t *testing.T) {
		h := HandlerKey("GET|")
		require.Equal(t, "", h.Path())
	})
}

func TestHandlerKey_RoundTrip(t *testing.T) {
	t.Run("Key() and HandlerKey round-trip preserves method and path", func(t *testing.T) {
		h := &CustomerHandler{Method: "DELETE", Path: "/api/v1/users/123"}
		key := HandlerKey(h.Key())
		require.Equal(t, "DELETE", key.Method())
		require.Equal(t, "/api/v1/users/123", key.Path())
	})

	t.Run("round-trip with path containing pipe", func(t *testing.T) {
		h := &CustomerHandler{Method: "GET", Path: "/api/v1/a|b"}
		key := HandlerKey(h.Key())
		require.Equal(t, "GET", key.Method())
		require.Equal(t, "/api/v1/a|b", key.Path())
	})
}
