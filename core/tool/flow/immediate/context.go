package immediate

import (
	"context"
	"time"
)

type Context struct {
	context.Context
	Params map[string]any
}

func WithCancel(ctx *Context) (*Context, context.CancelFunc) {
	c, cancel := context.WithCancel(ctx)
	return &Context{
		Context: c,
		Params:  ctx.Params,
	}, cancel
}

func WithTimeout(ctx *Context, timeout time.Duration) (*Context, context.CancelFunc) {
	c, cancel := context.WithTimeout(ctx, timeout)
	return &Context{
		Context: c,
		Params:  ctx.Params,
	}, cancel
}
