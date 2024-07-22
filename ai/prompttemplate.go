package ai

import "strings"

type PromptTemplate struct {
	template string
	sub      map[string]PromptTemplate
}

func (p PromptTemplate) Render(params map[string]string) string {
	if len(p.sub) > 0 {
		for k, v := range p.sub {
			params[k] = v.Render(params)
		}
	}

	replaceStr := make([]string, 0, len(params)*2)
	for k, v := range params {
		replaceStr = append(replaceStr, "{{#"+k+"#}}", v)
	}
	var replacer *strings.Replacer
	if len(replaceStr) > 0 {
		replacer = strings.NewReplacer(replaceStr...)
	} else {
		replacer = strings.NewReplacer()
	}

	return replacer.Replace(p.template)
}
