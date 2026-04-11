package tool

import (
	"context"
	"errors"
	"sync"
)

type AnyGroup struct {
	parent context.Context
	ctx    context.Context
	cancel context.CancelFunc
	funcs  []func(ctx context.Context) error
}

func WithContext(ctx context.Context) *AnyGroup {
	if ctx == nil {
		ctx = context.Background()
	}
	parent := ctx
	ctx, cancel := context.WithCancel(ctx)
	g := &AnyGroup{
		parent: parent,
		ctx:    ctx,
		cancel: cancel,
	}
	return g
}

func (g *AnyGroup) Go(fn func(ctx context.Context) error) {
	g.funcs = append(g.funcs, fn)
}

func (g *AnyGroup) Context() context.Context {
	return g.ctx
}

func (g *AnyGroup) checkContext() func() {
	if g.parent == nil {
		g.parent = context.Background()
	}
	if g.ctx == nil {
		g.ctx, g.cancel = context.WithCancel(g.parent)
	}
	return func() {
		g.cancel()
		g.ctx, g.cancel = context.WithCancel(g.parent)
	}
}

func (g *AnyGroup) WaitFirst() error {
	if len(g.funcs) == 0 {
		return nil
	}

	defer g.checkContext()()

	if g.parent.Err() != nil {
		return g.parent.Err()
	}

	funcs := g.funcs
	g.funcs = nil

	var wg sync.WaitGroup
	var errs []error
	var errLock sync.Mutex
	ch := make(chan struct{}, 1)
	defer close(ch)

	wg.Add(len(funcs))
	for _, fn := range funcs {
		go func(f func(ctx context.Context) error) {
			defer wg.Done()
			err := f(g.ctx)
			if err != nil {
				errLock.Lock()
				errs = append(errs, err)
				errLock.Unlock()
				return
			}
			select {
			case ch <- struct{}{}:
			default:
			}
		}(fn)
	}

	go func() {
		select {
		case <-g.ctx.Done():
			for range ch {
			}
			return
		case <-ch:
			g.cancel()
			return
		}
	}()

	wg.Wait()

	if len(errs) > 0 && len(errs) == len(funcs) {
		return errors.Join(errs...)
	}
	return nil
}

func (g *AnyGroup) WaitAll() error {
	if len(g.funcs) == 0 {
		return nil
	}

	defer g.checkContext()()

	if g.parent.Err() != nil {
		return g.parent.Err()
	}

	funcs := g.funcs
	g.funcs = nil

	var wg sync.WaitGroup
	var errs []error
	var errLock sync.Mutex

	wg.Add(len(funcs))
	for _, fn := range funcs {
		go func(f func(ctx context.Context) error) {
			defer wg.Done()
			err := f(g.ctx)
			if err != nil {
				errLock.Lock()
				errs = append(errs, err)
				errLock.Unlock()
			}
		}(fn)
	}

	wg.Wait()

	if len(errs) > 0 && len(errs) == len(funcs) {
		return errors.Join(errs...)
	}
	return nil
}
