package node

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/goslacker/slacker/extend/templatex"
)

func renderPrompt(tpl string, params map[string]any) (result string, err error) {
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
