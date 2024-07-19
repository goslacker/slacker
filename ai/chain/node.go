package chain

import "github.com/google/uuid"

type NodeInfo struct {
	id       string //uuid
	Name     string
	NextName string
}

func (m *NodeInfo) GetID() string {
	if m.id == "" {
		m.id = uuid.NewString()
	}
	return m.id
}

func (m NodeInfo) GetName() string {
	return m.Name
}

func (m NodeInfo) GetNextName() string {
	return m.NextName
}
