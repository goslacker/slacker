package ai

type Params interface {
	AfterNodeRun(nodes ...Node)
	Set(id, key string, value any)
	Get(partten string) any
}

type Node interface {
	GetID() string
	GetName() string
	GetNextName() string
	Run(params Params) (nextID string, err error) //map[nodeName]map[key]value
}

type Chain interface {
	AddNodes(nodes ...Node)
	Node
}
