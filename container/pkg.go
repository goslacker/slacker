package container

import (
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

func Bind[T any](providerOrInstance any, sets ...func(*bindOpts)) (err error) {
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

func Invoke(f any, opts ...func(*invokeOpts)) (err error) {
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
