package chain

import (
	"fmt"
	"github.com/goslacker/slacker/ai"
)

type Chain struct {
	NodeInfo
	nodes     map[string]ai.Node
	firstName string
}

func (c Chain) AddNodes(nodes ...ai.Node) {
	for _, node := range nodes {
		c.nodes[node.GetName()] = node
	}
	c.firstName = nodes[0].GetName()

	return
}

func (c Chain) Run(params map[string]map[string]any) (err error) {
	nodeName := c.firstName
	for {
		node := c.nodes[nodeName]
		err = node.Run(params)
		if err != nil {
			err = fmt.Errorf("run node <%s> failed: %w", node.GetName(), err)
			return
		}
	}
}
