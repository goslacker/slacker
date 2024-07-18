package container

import (
	"context"
	"fmt"
	"reflect"
)

type provider struct {
	Func      reflect.Value
	Singleton bool
}

type invokeOpts struct {
	params map[int]any
	keys   map[int]string
}

func WithParams(params map[int]any) func(*invokeOpts) {
	return func(opts *invokeOpts) {
		opts.params = params
	}
}

type bindOpts struct {
	Singleton bool
	Key       string
}

func NoSingleton() func(*bindOpts) {
	return func(opts *bindOpts) {
		opts.Singleton = false
	}
}

func WithKey(key string) func(*bindOpts) {
	return func(opts *bindOpts) {
		opts.Key = key
	}
}

func NewContainer() *Container {
	return &Container{
		providers: make(map[reflect.Type]map[string]*provider),
		instances: make(map[reflect.Type]map[string]reflect.Value),
	}
}

type Container struct {
	providers map[reflect.Type]map[string]*provider
	instances map[reflect.Type]map[string]reflect.Value
}

func (c *Container) Bind(t reflect.Type, value reflect.Value, opts ...func(*bindOpts)) (err error) {
	options := &bindOpts{
		Singleton: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	delete(c.instances, t)
	if canBindConsistent(t, value) {
		err = c.bindInstance(t, value, options)
	} else if canBindProvider(t, value) {
		err = c.bindProvider(t, value, options)
	} else {
		err = fmt.Errorf("can not bind <%s> to <%s>", value.Type(), t)
	}

	return
}

// bindConsistent 目标类型与给定值类型一致的情况
func (c *Container) bindInstance(t reflect.Type, value reflect.Value, options *bindOpts) (err error) {
	if t != value.Type() {
		value = value.Convert(t)
	}

	group, ok := c.instances[t]
	if !ok {
		group = make(map[string]reflect.Value)
		c.instances[t] = group
	}
	group[options.Key] = value

	return
}

func (c *Container) bindProvider(t reflect.Type, value reflect.Value, options *bindOpts) (err error) {
	group, ok := c.providers[t]
	if !ok {
		group = make(map[string]*provider)
		c.providers[t] = group
	}
	group[options.Key] = &provider{
		Func:      value,
		Singleton: options.Singleton,
	}

	return
}

func (c *Container) invoke(ctx context.Context, fn reflect.Value, opts ...func(opts *invokeOpts)) (rets []reflect.Value, err error) {
	options := &invokeOpts{
		params: make(map[int]any),
		keys:   make(map[int]string),
	}
	for _, opt := range opts {
		opt(options)
	}

	paramValues := make([]reflect.Value, 0, fn.Type().NumIn())
	waits := make([]reflect.Type, 0)
	for i := 0; i < fn.Type().NumIn(); i++ {
		var param reflect.Value
		if p, ok := options.params[i]; ok {
			param = reflect.ValueOf(p)
		} else {
			param, err = c.resolve(ctx, fn.Type().In(i), options.keys[i])
			if err != nil {
				return
			}
			if param.IsNil() {
				waits = append(waits, fn.Type().In(i))
			}
		}
		paramValues = append(paramValues, param)
	}
	results := fn.Call(paramValues)

	if len(results) == 0 {
		return
	}

	//check call error
	last := results[len(results)-1]
	e, ok := last.Interface().(error)
	if ok && e != nil {
		err = e
		return
	}

	chain := ctx.Value("chain").(*resolveChain)
	for _, typ := range waits {
		chain.Wait(typ, results[0])
	}

	if ok {
		rets = results[:len(results)-1]
	} else {
		rets = results
	}

	return
}

func (c *Container) resolve(ctx context.Context, t reflect.Type, key string) (ret reflect.Value, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Value("chain") == nil {
		ctx = context.WithValue(ctx, "chain", newResolveChain())
	}
	chain := ctx.Value("chain").(*resolveChain)
	defer func() {
		if !ret.IsValid() || ret.IsNil() {
			return
		}
		chain.FillWait(t, ret)
	}()

	//resolve in instances
	{
		if group, ok := c.instances[t]; ok {
			if ret, ok = group[key]; ok {
				return
			}
		}

		// if no target found, resolve witch can be converted to target
		for typ, group := range c.instances {
			if typ.ConvertibleTo(t) {
				var ok bool
				if ret, ok = group[key]; ok {
					ret = ret.Convert(t)
					return
				}
			}
		}
	}

	//resolve in providers
	{
		//check circular
		if chain.IsCircular(t) {
			ret = reflect.Zero(t)
			return
		}

		if group, ok := c.providers[t]; ok {
			if value, ok := group[key]; ok {
				var rets []reflect.Value
				rets, err = c.invoke(ctx, value.Func)
				if err != nil {
					return
				}
				ret = rets[0]
				if value.Singleton {
					if c.instances[t] == nil {
						c.instances[t] = make(map[string]reflect.Value)
					}
					c.instances[t][key] = ret
				}
				return
			}
		}

		// if no target found, resolve witch can be converted to target
		for typ, group := range c.providers {
			if typ.ConvertibleTo(t) {
				if prvd, ok := group[key]; ok {
					var rets []reflect.Value
					rets, err = c.invoke(ctx, prvd.Func)
					if err != nil {
						return
					}
					ret = rets[0].Convert(t)
					if prvd.Singleton {
						if c.instances[t] == nil {
							c.instances[t] = make(map[string]reflect.Value)
						}
						c.instances[t][key] = ret
					}
					return
				}
			}
		}
	}
	err = fmt.Errorf("type <%s> not found", t.String())
	return
}

func (c *Container) Resolve(t reflect.Type, key string) (result reflect.Value, err error) {
	return c.resolve(nil, t, key)
}

func (c *Container) Invoke(f reflect.Value, opts ...func(*invokeOpts)) (results []reflect.Value, err error) {
	return c.invoke(context.Background(), f, opts...)
}

func (c *Container) clear() {
	c.instances = make(map[reflect.Type]map[string]reflect.Value)
	c.providers = make(map[reflect.Type]map[string]*provider)
}

func canBindConsistent(t reflect.Type, value reflect.Value) bool {
	if t == value.Type() {
		return true
	}

	if value.Type().ConvertibleTo(t) {
		return true
	}

	return false
}

// canBindProvider
func canBindProvider(t reflect.Type, value reflect.Value) bool {
	if value.Kind() != reflect.Func {
		return false
	}

	if value.Type().NumOut() <= 0 {
		return false
	}

	resultType := value.Type().Out(0)

	if t == resultType {
		return true
	}

	if resultType.ConvertibleTo(t) {
		return true
	}

	return false
}
