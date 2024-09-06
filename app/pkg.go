package app

import (
	"github.com/goslacker/slacker/container"
	"github.com/goslacker/slacker/eventbus"
)

var app *App

func Default() *App {
	if app == nil {
		app = NewApp()
	}
	return app
}

// Use is shortcut of RegisterComponent
func Use(components ...Component) {
	RegisterComponent(components...)
}

func RegisterComponent(components ...Component) {
	Default().RegisterComponent(components...)
}

func Run() (n int, err error) {
	return Default().Run()
}
func Shutdown() {
	Default().Shutdown()
}
func RunAndWait() (err error) {
	return Default().RunAndWait()
}

func Bind[T any](providerOrInstance any, sets ...func(*container.BindOpts)) (err error) {
	return container.Bind[T](providerOrInstance, sets...)
}

func Resolve[T any](key ...string) (result T, err error) {
	return container.Resolve[T](key...)
}

func MustResolve[T any]() (result T) {
	return container.MustResolve[T]()
}

func Invoke(f any, opts ...func(*container.InvokeOpts)) (err error) {
	return container.Invoke(f, opts...)
}

func ResolveDirectly[T any](provider any, opts ...func(*container.InvokeOpts)) (result T, err error) {
	return container.ResolveDirectly[T](provider, opts...)
}

func MustResolveDirectly[T any](provider any, opts ...func(*container.InvokeOpts)) (result T) {
	return container.MustResolveDirectly[T](provider, opts...)
}

func RegisterListener[T any](listeners ...eventbus.ListenerFunc[T]) {
	eventbus.Register(listeners...)
}
func Fire[T any](event T) {
	eventbus.Fire(event)
}

func GetContainer() *container.Container {
	return container.Default()
}
