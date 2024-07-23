package chain

import (
	"context"
	"github.com/goslacker/slacker/ai"
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
	path   []string
	params map[string]any
	lock   sync.RWMutex
}

func (p *Ctx) AfterNodeRun(node Node) {
	p.path = append(p.path, node.GetID())
}

func (p *Ctx) SetParam(key string, value any) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.params[key] = value
}

func (p *Ctx) GetParam(key string) any {
	v, ok := p.params[key]
	if !ok {
		return nil
	} else {
		return v
	}
}

func (p *Ctx) GetParams(keys []string) map[string]any {
	ret := make(map[string]any, len(keys))
	for _, key := range keys {
		v := p.GetParam(key)
		if v != nil {
			ret[key] = v
		}
	}

	return ret
}

func NewChatCtx() ChatContext {
	return &ChatCtx{
		Context: NewContext(context.Background()),
		History: NewHistory(),
	}
}

type ChatCtx struct {
	Context
	*History
}

func (c ChatCtx) SetHistory(messages ...ai.Message) {
	c.History.Set(messages...)
}

func (c ChatCtx) GetHistory(limit int) (messages []ai.Message) {
	return c.History.Get(limit)
}

func WithHistory(parent Context, history *History) Context {
	switch x := parent.(type) {
	case *ChatCtx:
		x.History = history
		return x
	default:
		return ChatCtx{
			Context: parent,
			History: history,
		}
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
