package chain

import (
	"context"
	"github.com/goslacker/slacker/ai/client"
)

type Context interface {
	context.Context
	SetParam(key string, value any)
	GetParam(pattern string) any
	GetAllParams() map[string]any
	SetHistory(messages ...client.Message)
	GetHistory(limit int) (messages []client.Message)
}

type Node interface {
	GetID() string
	GetName() string
	Run(ctx Context) (nextID string, err error)
}

type Chain interface {
	AddNodes(nodes ...Node)
	Node
}
