package httpx

import "strings"

type URL string

func (u URL) Append(uri string) string {
	return strings.Join([]string{strings.TrimRight(string(u), "/"), strings.TrimLeft(uri, "/")}, "/")
}
