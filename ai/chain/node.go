package chain

import (
	"github.com/google/uuid"
)

type NodeInfo struct {
	id string //uuid
}

func (m *NodeInfo) GetID() string {
	if m.id == "" {
		m.id = uuid.NewString()
	}
	return m.id
}

func WithNextID(id string) func(node *LLMNode) {
	return func(node *LLMNode) {
		node.nextID = id
	}
}

func WithParamKeys(paramKeys ...string) func(node *LLMNode) {
	return func(node *LLMNode) {
		node.paramKeys = paramKeys
	}
}

func NewLLMNode(opts ...func(node *LLMNode)) *LLMNode {
	n := &LLMNode{}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

type LLMNode struct {
	NodeInfo
	nextID    string
	paramKeys []string
}
