package eventbus

import (
	"reflect"
	"sync"
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
	var wg sync.WaitGroup
	wg.Add(len(listeners))
	for _, listener := range listeners {
		go func() {
			defer wg.Done()
			listener.Call([]reflect.Value{reflect.ValueOf(event)})
		}()
	}
	wg.Wait()
}
