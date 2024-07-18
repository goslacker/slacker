package container

import (
	"github.com/goslacker/slacker/extend/reflectx"
	"reflect"
)

func newResolveChain() *resolveChain {
	return &resolveChain{
		chain: make([]reflect.Type, 0),
		waits: make(map[reflect.Type][]reflect.Value),
	}
}

type resolveChain struct {
	chain []reflect.Type
	waits map[reflect.Type][]reflect.Value
}

func (r *resolveChain) IsCircular(t reflect.Type) bool {
	if len(r.chain) == 0 {
		r.chain = append(r.chain, t)
		return false
	}

	for i := len(r.chain) - 1; i >= 0; i-- {
		if t == r.chain[i] {
			return true
		}
	}

	r.chain = append(r.chain, t)
	return false
}

func (r resolveChain) Wait(t reflect.Type, v reflect.Value) (ret reflect.Value) {
	r.waits[t] = append(r.waits[t], v)
	return
}

func (r *resolveChain) HasWait(t reflect.Type) bool {
	_, ok := r.waits[t]
	return ok
}

func (r *resolveChain) FillWait(t reflect.Type, v reflect.Value) {
	if !r.HasWait(t) {
		return
	}
	if s, ok := r.waits[t]; ok {
		for _, item := range s {
			item = reflectx.Indirect(item, false)
			for i := 0; i < item.NumField(); i++ {
				field := item.Field(i)
				if t == field.Type() {
					target := reflect.NewAt(t, field.Addr().UnsafePointer())
					target.Elem().Set(v)
				}
			}
		}
	}
	for {
		tmp := r.chain[len(r.chain)-1]
		r.chain = r.chain[:len(r.chain)-1]
		if tmp == t || len(r.chain) == 0 {
			break
		}
	}
}
