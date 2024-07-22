package chain

import (
	"context"
	"github.com/goslacker/slacker/ai"
	"strings"
	"sync"
	"time"
)

func NewContext(ctx context.Context) Context {
	return &Ctx{
		Context: ctx,
		params:  make(map[string]map[string]any),
	}
}

type Ctx struct {
	context.Context
	*History
	path   []string
	params map[string]map[string]any
	lock   sync.RWMutex
}

func (p *Ctx) AfterNodeRun(node Node) {
	p.path = append(p.path, node.GetID())
}

func (p *Ctx) SetParam(id, key string, value any) {
	p.lock.Lock()
	defer p.lock.Unlock()
	group, ok := p.params[id]
	if !ok {
		group = make(map[string]any)
		p.params[id] = group
	}
	group[key] = value
}

func (p *Ctx) GetParam(pattern string) any {
	paramPath := strings.Split(pattern, "/")
	group, ok := p.params[paramPath[0]]
	if !ok {
		return nil
	}
	return group[paramPath[1]]
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
