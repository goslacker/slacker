package chain

import (
	"fmt"
)

func NewChain() Chain {
	return &chain{
		nodes: make(map[string]Node),
	}
}

type chain struct {
	NodeInfo
	nodes   map[string]Node
	FirstID string
	LastID  string
}

func (c *chain) AddNodes(nodes ...Node) {
	for _, node := range nodes {
		c.nodes[node.GetID()] = node
	}
	c.FirstID = nodes[0].GetID()
	c.LastID = nodes[len(nodes)-1].GetID()

	return
}

func (c *chain) Run(ctx Context) (nextID string, err error) {
	nodeID := c.FirstID
	for {
		node := c.nodes[nodeID]
		nodeID, err = node.Run(ctx)
		if err != nil {
			err = fmt.Errorf("run node <%s> failed: %w", node.GetID(), err)
			return
		}
		ctx.AfterNodeRun(node)
		if node.GetID() == c.LastID {
			return
		}
	}
}

func NewChatChain() *ChatChain {
	return &ChatChain{
		Chain:   NewChain(),
		History: NewHistory(),
	}
}

type ChatChain struct {
	Chain
	*History
}

func (c *ChatChain) AddNodes(nodes ...Node) {
	for _, node := range nodes {
		if n, ok := node.(CanSetMessageHistory); ok {
			n.SetMessageHistorySetter(c.History.Set)
		}
		if n, ok := node.(CanGetMessageHistory); ok {
			n.SetMessageHistoryGetter(c.History.Get)
		}
	}
	c.Chain.AddNodes(nodes...)
}
