package ruleengine

import "sync"

type variableKey string

var VariableKey variableKey = "variable"

type Variable struct {
	m    map[string]string
	lock sync.RWMutex
}

func (v *Variable) Set(key, value string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.m[key] = value
}

func (v *Variable) Get(key string) string {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.m[key]
}