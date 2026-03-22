package eventbus

import (
	"reflect"
)

type ListenerFunc[T any] func(event T) error

var listenerMap = make(map[reflect.Type][]reflect.Value)

func Register[T any](listeners ...ListenerFunc[T]) {
	for _, listener := range listeners {
		v := reflect.ValueOf(listener)
		eventType := v.Type().In(0)
		listenerMap[eventType] = append(listenerMap[eventType], v)
	}
}

func Fire[T any](event T) (err error) {
	t := reflect.TypeOf(event)
	listeners, ok := listenerMap[t]
	if !ok {
		return
	}
	for _, listener := range listeners {
		results := listener.Call([]reflect.Value{reflect.ValueOf(event)})
		if !results[0].IsNil() {
			err = results[0].Interface().(error)
			return
		}
	}
	return
}
