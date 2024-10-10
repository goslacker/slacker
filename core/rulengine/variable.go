package ruleengine

import "sync"

type variableKey string

var VariableKey variableKey = "variable"

func NewVariable() *Variable {
	return &Variable{
		m: make(map[string]any),
	}
}

type Variable struct {
	m    map[string]any
	lock sync.RWMutex
}

func (v *Variable) Set(key string, value any) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.m[key] = value
}

func (v *Variable) Get(key string) (result any, ok bool) {
	v.lock.RLock()
	defer v.lock.RUnlock()
	result, ok = v.m[key]
	return
}
