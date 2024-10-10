package templatex

import (
	"bytes"
	"fmt"
	"text/template"
)

func NewTextTemplate(template string) *TextTemplate {
	return &TextTemplate{
		template: template,
	}
}

type TextTemplate struct {
	subTemplates map[string]*TextTemplate
	template     string
}

func (t *TextTemplate) AddSub(key string, subTemplate *TextTemplate) {
	if t.subTemplates == nil {
		t.subTemplates = make(map[string]*TextTemplate)
	}
	t.subTemplates[key] = subTemplate
}

func (t *TextTemplate) RenderWithFuncMap(data map[string]any, funcMap template.FuncMap) (result string, err error) {
	for key, subTemplate := range t.subTemplates {
		data[key], err = subTemplate.RenderWithFuncMap(data, funcMap)
		if err != nil {
			err = fmt.Errorf("render sub template <%s> failed: %w", key, err)
			return
		}
	}

	temp := template.New("template")
	if len(funcMap) > 0 {
		temp.Funcs(funcMap)
	}

	temp, err = temp.Parse(t.template)
	if err != nil {
		err = fmt.Errorf("parse template failed: %w", err)
		return
	}

	r := bytes.NewBuffer(make([]byte, 0, len(t.template)))
	err = temp.Option("missingkey=error").Execute(r, data)
	if err != nil {
		err = fmt.Errorf("exec template failed: %w", err)
		return
	}
	result = r.String()
	return
}

func (t *TextTemplate) Render(data map[string]any) (result string, err error) {
	return t.RenderWithFuncMap(data, nil)
}
