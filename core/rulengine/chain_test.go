package ruleengine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testNode struct {
	ID string
}

func (t *testNode) GetID() string {
	return t.ID
}

func (t *testNode) Run(ctx context.Context) {
	if t.ID == "3" {
		time.Sleep(1 * time.Second)
	}
	println(t.ID)
	ctx.Value(ChainKey).(*Chain).Next(ctx, t)
}

type testEdge struct {
	source string
	target string
}

func (t *testEdge) GetSource() string {
	return t.source
}

func (t *testEdge) GetTarget() string {
	return t.target
}

func TestChain(t *testing.T) {
	t1 := &testNode{ID: "1"}
	t2 := &testNode{ID: "2"}
	t3 := &testNode{ID: "3"}
	t4 := &testNode{ID: "4"}
	t5 := &testNode{ID: "5"}

	e1 := &testEdge{source: "1", target: "2"}
	e2 := &testEdge{source: "1", target: "3"}
	e3 := &testEdge{source: "3", target: "4"}
	e4 := &testEdge{source: "4", target: "5"}
	e5 := &testEdge{source: "2", target: "5"}

	c := Chain{
		nodes: []Node{t1, t2, t3, t4, t5},
		edges: []Edge{e1, e2, e3, e4, e5},
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// defer cancel()
	c.Run(context.Background())
}

func TestSubChain(t *testing.T) {
	t1 := &testNode{ID: "1"}
	t2 := &testNode{ID: "2"}
	t3 := &testNode{ID: "3"}
	t4 := &testNode{ID: "4"}
	t5 := &testNode{ID: "5"}

	e1 := &testEdge{source: "1", target: "2"}
	e2 := &testEdge{source: "1", target: "3"}
	e3 := &testEdge{source: "3", target: "4"}
	e4 := &testEdge{source: "4", target: "5"}
	e5 := &testEdge{source: "2", target: "5"}

	c := Chain{
		ID:    "sub",
		nodes: []Node{t1, t2, t3, t4, t5},
		edges: []Edge{e1, e2, e3, e4, e5},
	}

	tstart := &testNode{ID: "start"}
	tstop := &testNode{ID: "stop"}

	parentC := Chain{
		nodes: []Node{tstart, tstop, &c},
		edges: []Edge{
			&testEdge{source: "start", target: "sub"},
			&testEdge{source: "sub", target: "stop"},
		},
	}
	parentC.Run(context.Background())
}

type testDefaultNode struct {
	DefaultNode
}

func (t *testDefaultNode) Run(ctx context.Context) {
	t.DefaultNode.WithRunFunc(func(ctx context.Context, params map[string]any) (result map[string]any, err error) {
		if t.ID != "stop" {
			params["a"] = params["a"].(int) + 1
		}
		return params, nil
	}).Run(ctx)
}

func TestParamNode(t *testing.T) {
	start := &testDefaultNode{
		DefaultNode: DefaultNode{
			ID:       "start",
			Name:     "start",
			ParamMap: map[string]string{"a": "start"},
		},
	}
	n1 := &testDefaultNode{
		DefaultNode: DefaultNode{
			ID:       "1",
			Name:     "1",
			ParamMap: map[string]string{"a": "start/a"},
		},
	}
	n2 := &testDefaultNode{
		DefaultNode: DefaultNode{
			ID:       "2",
			Name:     "2",
			ParamMap: map[string]string{"a": "1/a"},
		},
	}
	n3 := &testDefaultNode{
		DefaultNode: DefaultNode{
			ID:       "3",
			Name:     "3",
			ParamMap: map[string]string{"a": "2/a"},
		},
	}
	stop := &testDefaultNode{
		DefaultNode: DefaultNode{
			ID:       "stop",
			Name:     "stop",
			ParamMap: map[string]string{"a": "3/a"},
		},
	}
	c := Chain{
		nodes: []Node{start, n1, n2, n3, stop},
		edges: []Edge{
			&testEdge{source: "start", target: "1"},
			&testEdge{source: "1", target: "2"},
			&testEdge{source: "2", target: "3"},
			&testEdge{source: "3", target: "stop"},
		},
	}

	param := &Param{}
	param.SetWithPrefix(map[string]any{"start": 0}, "")
	ctx := context.WithValue(context.Background(), ParamKey, param)

	c.Run(ctx)
	require.Equal(t, 4, param.LoadByMap(map[string]string{"result": "stop/a"})["result"])
}
