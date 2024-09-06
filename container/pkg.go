package container

import (
	"errors"
	"reflect"
	"sync"
)

var lock sync.Mutex
var def *Container

func Set(container *Container) {
	lock.Lock()
	defer lock.Unlock()
	def = container
}

func Default() *Container {
	lock.Lock()
	defer lock.Unlock()
	if def == nil {
		def = NewContainer()
	}
	return def
}

func Bind[T any](providerOrInstance any, sets ...func(*BindOpts)) (err error) {
	return Default().Bind(reflect.TypeOf((*T)(nil)).Elem(), reflect.ValueOf(providerOrInstance), sets...)
}

func Resolve[T any](key ...string) (result T, err error) {
	var k string
	if len(key) > 0 {
		k = key[0]
	}
	res, err := Default().Resolve(reflect.TypeOf((*T)(nil)).Elem(), k)
	if err != nil {
		return
	}
	result = res.Interface().(T)
	return
}

func MustResolve[T any]() (result T) {
	result, err := Resolve[T]()
	if err != nil {
		panic(err)
	}
	return
}

func Invoke(f any, opts ...func(*InvokeOpts)) (err error) {
	results, err := Default().Invoke(reflect.ValueOf(f), opts...)
	if err != nil {
		return
	}
	if len(results) > 0 {
		e := results[len(results)-1].Interface()
		if e != nil {
			err = e.(error)
		}
	}
	return
}

func ResolveDirectly[T any](provider any, opts ...func(*InvokeOpts)) (result T, err error) {
	vp := reflect.ValueOf(provider)
	if vp.Type().NumOut() > 2 || vp.Type().NumOut() < 1 {
		err = errors.New("provider must return 1 or 2(with error) values")
		return
	}
	results, err := Default().Invoke(vp, opts...)
	if err != nil {
		return
	}
	if len(results) == 2 {
		if e := results[len(results)-1].Interface(); e != nil {
			err = e.(error)
			return
		}
	}
	result = results[0].Interface().(T)
	return
}

func MustResolveDirectly[T any](provider any, opts ...func(*InvokeOpts)) (result T) {
	result, err := ResolveDirectly[T](provider, opts...)
	if err != nil {
		panic(err)
	}
	return
}
