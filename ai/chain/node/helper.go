package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goslacker/slacker/ai/client"
	"github.com/goslacker/slacker/extend/templatex"
	"regexp"
	"strings"
)

func renderPrompt(tpl string, params map[string]any, history []client.Message) (result string, err error) {
	if strings.Contains(tpl, "{{#history#}}") {
		var tmp []byte
		tmp, err = json.Marshal(history)
		if err != nil {
			return
		}
		h := string(tmp)
		if h == "null" {
			params["history"] = ""
		} else {
			params["history"] = h
		}
	}

	reg := regexp.MustCompile(`\{\{#.*?#\}\}`)
	tpl = string(reg.ReplaceAllFunc([]byte(tpl), func(b []byte) []byte {
		b = bytes.TrimPrefix(b, []byte("{{#"))
		b = bytes.TrimSuffix(b, []byte("#}}"))
		b = bytes.TrimSpace(b)
		return []byte(fmt.Sprintf(`{{ index . "%s" }}`, string(b)))
	}))

	t := templatex.NewTextTemplate(tpl)
	return t.Render(params)
}
