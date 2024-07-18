package chain

type NodeInfo struct {
	Name     string
	NextName string
}

func (m NodeInfo) GetName() string {
	return m.Name
}

func (m NodeInfo) GetNextName() string {
	return m.NextName
}
