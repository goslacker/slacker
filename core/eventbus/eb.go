package eventbus

import (
	"reflect"
)

type ListenerFunc[T any] func(event T)

var listenerMap = make(map[reflect.Type][]reflect.Value)

func Register[T any](listeners ...ListenerFunc[T]) {
	for _, listener := range listeners {
		v := reflect.ValueOf(listener)
		listenerMap[v.Type().In(0)] = append(listenerMap[v.Type().In(0)], v)
	}

	return
}

func Fire[T any](event T) {
	t := reflect.TypeOf(event)
	listeners, ok := listenerMap[t]
	if !ok {
		return
	}
	ch := make(chan struct{})
	go func() {
		defer func() { ch <- struct{}{} }()
		for _, listener := range listeners {
			listener.Call([]reflect.Value{reflect.ValueOf(event)})
		}
	}()
	<-ch
	close(ch)
}
