package database

import (
	"github.com/goslacker/slacker/core/tool"
)

func DefaultM2E[PO any, Entity any](dst *Entity, src *PO) error {
	return tool.SimpleMap(dst, src)
}

func DefaultE2M[PO any, Entity any](dst *PO, src *Entity) error {
	return tool.SimpleMap(dst, src)
}

type txKey string

var TxKey txKey = "tx"
