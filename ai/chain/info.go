package chain

import (
	"github.com/google/uuid"
	"github.com/goslacker/slacker/extend/slicex"
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

type Info struct {
	ID        string
	Name      string
	Type      Type
	Variables []Variable
}

func (n *Info) VariableNames() []string {
	return slicex.Map(n.Variables, func(item Variable) string {
		return item.Name
	})
}

func (n *Info) GetID() string {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	return n.ID
}

func (n *Info) GetName() string {
	return n.Name
}

type Variable struct {
	Label string
	Name  string
	Type  VariableType
}
