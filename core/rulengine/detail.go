package ruleengine

import (
	"encoding/json"
	"log/slog"
	"sync"
)

type detailKey string

var DetailKey detailKey = "detail"

type detail struct {
	key   string
	value any
}

type Detail struct {
	details []any
	lock    sync.Mutex
}

func (d *Detail) Push(key string, value any) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.details = append(d.details, detail{key: key, value: value})
}

func (d *Detail) String() string {
	b, err := json.Marshal(d.details)
	if err != nil {
		slog.Error("marshal detail error", "err", err)
		return ""
	} else {
		return string(b)
	}
}
