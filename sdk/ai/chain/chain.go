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
	firstID string
	lastID  string
	nextID  string
}

func (c *chain) AddNodes(nodes ...Node) {
	for _, nod := range nodes {
		c.nodes[nod.GetID()] = nod
	}
	c.firstID = nodes[0].GetID()
	c.lastID = nodes[len(nodes)-1].GetID()

	return
}

func (c *chain) Run(ctx Context) (nextID string, err error) {
	nodeID := c.firstID
	for {
		nod := c.nodes[nodeID]
		nodeID, err = nod.Run(ctx)
		if err != nil {
			err = fmt.Errorf("run node %s<%s> failed: %w", nod.GetName(), nod.GetID(), err)
			return
		}
		if nod.GetID() == c.lastID {
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

func (c *ChatChain) Run(ctx Context) (nextID string, err error) {
	ctx.SetHistoryManager(c.History)
	return c.Chain.Run(ctx)
}
