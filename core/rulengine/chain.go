package ruleengine

import (
	"context"
	"errors"
	"github.com/goslacker/slacker/core/slicex"
	"sync"
)

type chainKey string

var ChainKey chainKey = "chain"

type Edge interface {
	GetTarget() string
	GetSource() string
}

func NewChain(ID string) *Chain {
	return &Chain{
		ID:    ID,
		nodes: make([]Node, 0, 10),
		edges: make([]Edge, 0, 10),
	}
}

type Chain struct {
	ID         string
	nodes      []Node
	edges      []Edge
	waits      sync.Map
	completeds sync.Map
	wg         sync.WaitGroup
	cancel     context.CancelFunc
}

func (c *Chain) GetID() string {
	if c.ID == "" {
		panic("chain's id is empty")
	}
	return c.ID
}

// Stop 停止规则链执行
func (c *Chain) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Chain) Next(ctx context.Context, current Node) {
	select {
	case <-ctx.Done():
		return
	default:
		if current == nil {
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				c.first().Run(ctx)
			}()
		} else {
			c.complete(current.GetID())
			nexts := c.findNexts(current)

			if len(nexts) == 0 {
				return
			}

			for _, next := range nexts {
				select {
				case <-ctx.Done():
					return
				default:
					if c.readyToRun(next.GetID()) {
						c.wg.Add(1)
						go func() {
							defer c.wg.Done()
							next.Run(ctx)
						}()
					}
				}
			}
		}
	}
}

func (c *Chain) SetContent(nodes []Node, edges []Edge) {
	c.nodes = nodes
	c.edges = edges
}

func (c *Chain) readyToRun(nextID string) bool {
	//先查next是不是正在等待
	waits, ok := c.waits.Load(nextID)
	if ok {
		tmp := slicex.Filter(waits.([]string), func(wait string) bool {
			return !c.isCompletee(wait)
		})
		if len(tmp) <= 0 {
			c.waits.Delete(nextID)
			return true
		} else {
			c.waits.Store(nextID, tmp)
			return false
		}
	}

	//再查前面的节点是否都完成了,没有完成就加到waits中
	edges := slicex.Filter(c.edges, func(edge Edge) bool {
		return edge.GetTarget() == nextID
	})

	if len(edges) <= 0 {
		return true
	} else {
		tmp := slicex.Filter(edges, func(edge Edge) bool {
			return !c.isCompletee(edge.GetSource())
		})
		if len(tmp) <= 0 {
			return true
		} else {
			c.addToWaits(nextID, slicex.Map(tmp, func(edge Edge) string {
				return edge.GetSource()
			})...)
			return false
		}
	}
}

func (c *Chain) isCompletee(nodeID string) bool {
	record, ok := c.completeds.Load(nodeID)
	return ok && record.(bool)
}

func (c *Chain) complete(nodeID string) {
	c.completeds.Store(nodeID, true)
}

func (c *Chain) addToWaits(nodeID string, waits ...string) {
	group, ok := c.waits.Load(nodeID)
	if !ok {
		group = waits
	} else {
		group = append(group.([]string), waits...)
	}
	c.waits.Store(nodeID, group)
}

func (c *Chain) findNexts(current Node) []Node {
	edges := slicex.Filter(c.edges, func(edge Edge) bool {
		return edge.GetSource() == current.GetID()
	})
	var nodes []Node
	for _, edge := range edges {
		node, ok := slicex.Find(c.nodes, func(node Node) bool {
			return node.GetID() == edge.GetTarget()
		})
		if ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (c *Chain) Run(ctx context.Context) {
	parent := ctx.Value(ChainKey)
	defer func() { //TODO: 添加错误处理
		if parent != nil {
			parent.(*Chain).Next(ctx, c)
		}
	}()
	if ctx.Value(DetailKey) == nil {
		ctx = context.WithValue(ctx, DetailKey, &Detail{})
	}
	if ctx.Value(VariableKey) == nil {
		ctx = context.WithValue(ctx, VariableKey, &Variable{})
	}
	ctx, c.cancel = context.WithCancel(ctx)
	ctx = context.WithValue(ctx, ChainKey, c)
	c.Next(ctx, nil)
	c.wg.Wait()
}

func (c *Chain) first() Node {
	first, ok := slicex.Find(c.nodes, func(node Node) bool {
		return !slicex.Contains(node.GetID(), slicex.Map(c.edges, func(edge Edge) string {
			return edge.GetTarget()
		}))
	})
	if !ok {
		panic(errors.New("first node not found"))
	}
	return first
}
