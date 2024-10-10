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

func WithM2E[PO any, Entity any](f func(dst *Entity, src *PO) error) func(Repository[Entity]) {
	return func(r Repository[Entity]) {
		r.SetM2E(f)
	}
}

func WithE2M[PO any, Entity any](f func(dst *PO, src *Entity) error) func(Repository[Entity]) {
	return func(r Repository[Entity]) {
		r.SetE2M(f)
	}
}

type txKey string

var TxKey txKey = "tx"
