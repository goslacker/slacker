package chain

import (
	"context"
	"github.com/goslacker/slacker/sdk/ai/client"
)

type Context interface {
	context.Context
	SetParam(key string, value any)
	GetParam(pattern string) any
	GetAllParams() map[string]any
	SetHistory(messages ...client.Message)
	GetHistory(limit int) (messages []client.Message)
	SetHistoryManager(history *History)
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
