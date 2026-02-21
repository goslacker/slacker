package grpcgatewayx

import "net/http"

func code2http(code int32) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

var codes = map[int32]int{}

func RegisterCode(code int32, httpCode int) {
	codes[code] = httpCode
}
