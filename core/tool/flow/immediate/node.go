package immediate

import "context"

// Node is a node in the chain.
type Node interface {
	Run(ctx context.Context) (err error)
	Next() Node
}

// Input is the input of a node.
type Inputable interface {
	SetInputParams(Params) map[string]any
	GetInputParams() map[string]any
}

// Output is the output of a node.
type Outputable interface {
	SetOutputParams(Params)
}

type NormalNode struct {
	input Inputable
}
