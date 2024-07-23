package chain

import (
	"github.com/google/uuid"
)

type NodeInfo struct {
	ID string //uuid
}

func (m *NodeInfo) GetID() string {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return m.ID
}
