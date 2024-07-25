package urlx

import "strings"

func ParseQuery(m map[string][]string, key string) (map[string][]string, bool) {
	dicts := make(map[string][]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v
			}
		}
	}
	return dicts, exist
}
