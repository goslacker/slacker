package chain

import (
	"github.com/google/uuid"
	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/core/tool"
)

type VariableType string

const (
	TypeString VariableType = "string"
	TypeNumber VariableType = "number"
)

type Type string

const (
	TypeLLM   Type = "llm"
	TypeStart Type = "start"
	TypeEnd   Type = "end"
)

type NodeInfo struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      Type       `json:"type"`
	Variables []Variable `json:"variables"`
}

func (n *NodeInfo) VariableNames() []string {
	return slicex.Map(n.Variables, func(item Variable) string {
		return item.Name
	})
}

func (n *NodeInfo) GetID() string {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	return n.ID
}

func (n *NodeInfo) GetName() string {
	return tool.Ternary(n.Name == "", n.GetID(), n.Name)
}

type Variable struct {
	Label string       `json:"label"`
	Name  string       `json:"name"`
	Type  VariableType `json:"type"`
}
