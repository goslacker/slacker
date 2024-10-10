package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func NewRequest(method string, uri string, data any, marshal ...func(data any) ([]byte, error)) (req *http.Request, err error) {
	var body []byte
	switch x := data.(type) {
	case []byte:
		body = x
	case string:
		body = []byte(x)
	default:
		m := json.Marshal
		if len(marshal) > 0 {
			m = marshal[0]
		}
		body, err = m(data)
		if err != nil {
			return
		}
	}

	return http.NewRequest(method, uri, bytes.NewReader(body))
}
