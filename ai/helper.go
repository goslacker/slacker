package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goslacker/slacker/extend/templatex"
	"regexp"
)

func RenderPrompt(tpl string, params map[string]any, history []Message) (result string, err error) {
	reg := regexp.MustCompile(`\{\{#.*?#\}\}`)
	tpl = string(reg.ReplaceAllFunc([]byte(tpl), func(b []byte) []byte {
		b = bytes.TrimPrefix(b, []byte("{{#"))
		b = bytes.TrimSuffix(b, []byte("#}}"))
		b = bytes.TrimSpace(b)
		return []byte(fmt.Sprintf(`{{ index . "%s" }}`, string(b)))
	}))

	t := templatex.NewTextTemplate(tpl)
	tmp, err := json.Marshal(history)
	if err != nil {
		return
	}
	params["history"] = string(tmp)

	return t.Render(params)
}
