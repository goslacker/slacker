package chain

import (
	"fmt"
	"github.com/goslacker/slacker/ai"
)

func NewChain() *Chain {
	return &Chain{
		nodes: make(map[string]ai.Node),
	}
}

type Chain struct {
	NodeInfo
	nodes   map[string]ai.Node
	FirstID string
	LastID  string
}

func (c *Chain) AddNodes(nodes ...ai.Node) {
	for _, node := range nodes {
		c.nodes[node.GetID()] = node
	}
	c.FirstID = nodes[0].GetID()
	c.LastID = nodes[len(nodes)-1].GetID()

	return
}

func (c *Chain) Run(params ai.Params) (nextID string, err error) {
	nodeID := c.FirstID
	for {
		node := c.nodes[nodeID]
		nodeID, err = node.Run(params)
		if err != nil {
			err = fmt.Errorf("run node <%s> failed: %w", node.GetName(), err)
			return
		}
		params.AfterNodeRun(node)
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
	*Chain
	*History
}

func (c *ChatChain) AddNodes(nodes ...ai.Node) {
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
