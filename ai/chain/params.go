package chain

import (
	"github.com/goslacker/slacker/ai"
	"strings"
	"sync"
)

func NewParams() *Params {
	return &Params{
		params: make(map[string]map[string]any),
	}
}

type Params struct {
	Path   []string
	params map[string]map[string]any
	lock   sync.RWMutex
}

func (p *Params) AfterNodeRun(node ai.Node) {
	p.Path = append(p.Path, node.GetID())
}

func (p *Params) Set(id, key string, value any) {
	p.lock.Lock()
	defer p.lock.Unlock()
	group, ok := p.params[id]
	if !ok {
		group = make(map[string]any)
		p.params[id] = group
	}
	group[key] = value
}

func (p *Params) Get(pattern string) any {
	paramPath := strings.Split(pattern, "/")
	group, ok := p.params[paramPath[0]]
	if !ok {
		return nil
	}
	return group[paramPath[1]]
}
