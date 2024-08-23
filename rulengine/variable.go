package ruleengine

import "sync"

type variableKey string

var VariableKey variableKey = "variable"

type Variable struct {
	m    map[string]any
	lock sync.RWMutex
}

func (v *Variable) Set(key string, value any) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.m[key] = value
}

func (v *Variable) Get(key string) any {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.m[key]
}
