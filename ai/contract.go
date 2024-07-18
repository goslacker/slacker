package ai

type Node interface {
	GetName() string
	GetNextName() string
	Run(params map[string]map[string]any) (err error) //map[nodeName]map[key]value
}

type Chain interface {
	AddNodes(nodes ...Node)
	Node
}
