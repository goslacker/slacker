package chain

import (
	"context"
	"github.com/goslacker/slacker/ai"
)

type Context interface {
	context.Context
	AfterNodeRun(node Node)
	SetParam(id, key string, value any)
	GetParam(pattern string) any
}

type ChatContext interface {
	Context
	ChatHistoryManager
}

type ChatHistoryManager interface {
	SetHistory(messages ...ai.Message)
	GetHistory(limit int) (messages []ai.Message)
}

type Node interface {
	GetID() string
	Run(ctx Context) (nextID string, err error) //map[nodeName]map[key]value
}

type Chain interface {
	AddNodes(nodes ...Node)
	Node
}
