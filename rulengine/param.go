package ruleengine

import "sync"

type paramKey string

var ParamKey paramKey = "param"

type Param struct {
	m sync.Map
}

func (p *Param) SetWithPrefix(m map[string]any, prefix string) {
	for key, value := range m {
		if prefix != "" {
			key = prefix + "/" + key
		}
		p.m.Store(key, value)
	}
}

func (p *Param) LoadByMap(paramMap map[string]string) map[string]any {
	result := make(map[string]any, len(paramMap))
	for mapKey, sourceKey := range paramMap {
		if value, ok := p.m.Load(sourceKey); ok {
			result[mapKey] = value
		}
	}

	return result
}

func (p *Param) Get(key string) any {
	ret, _ := p.m.Load(key)
	return ret
}
