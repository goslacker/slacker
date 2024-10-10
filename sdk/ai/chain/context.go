package chain

import (
	"context"
	"github.com/goslacker/slacker/sdk/ai/client"
	"sync"
	"time"
)

func NewContext(ctx context.Context) Context {
	return &Ctx{
		Context: ctx,
		params:  make(map[string]any),
	}
}

type Ctx struct {
	context.Context
	params  map[string]any
	lock    sync.RWMutex
	history *History
}

func (c *Ctx) SetParam(key string, value any) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.params[key] = value
}

func (c *Ctx) GetParam(key string) any {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v, ok := c.params[key]
	if !ok {
		return nil
	} else {
		return v
	}
}

func (c *Ctx) GetAllParams() map[string]any {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.params
}

func (c *Ctx) SetHistory(messages ...client.Message) {
	c.history.Set(messages...)
}

func (c *Ctx) GetHistory(limit int) (messages []client.Message) {
	return c.history.Get(limit)
}

func (c *Ctx) SetHistoryManager(history *History) {
	c.history = history
}

func WithHistory(parent Context, history *History) Context {
	return &Ctx{
		params:  parent.GetAllParams(),
		Context: parent,
		history: history,
	}
}

func WithCancel(parent Context) (Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	c := NewContext(ctx)
	return c, cancel
}

func WithTimeout(parent Context, timeout time.Duration) (Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	c := NewContext(ctx)
	return c, cancel
}
