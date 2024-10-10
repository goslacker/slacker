package ruleengine

import (
	"context"
	"fmt"
	"log/slog"
)

type Node interface {
	GetID() string
	Run(ctx context.Context)
}

type DefaultNode struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	ParamMap   map[string]string `json:"paramMap"`
	runFunc    func(ctx context.Context, params map[string]any) (result map[string]any, err error)
	detailFunc func(params map[string]any, result map[string]any) any
}

func (n *DefaultNode) GetID() string {
	return n.ID
}

func (n *DefaultNode) loadParam(ctx context.Context) map[string]any {
	param := make(map[string]any, len(n.ParamMap))
	if len(n.ParamMap) > 0 {
		paramManager := ctx.Value(ParamKey)
		if paramManager == nil {
			panic("param manager not found")
		}
		param = paramManager.(*Param).LoadByMap(n.ParamMap)
	}
	return param
}

func (n *DefaultNode) setParam(ctx context.Context, param map[string]any) {
	paramManager := ctx.Value(ParamKey)
	if paramManager == nil {
		panic("param manager not found")
	}
	paramManager.(*Param).SetWithPrefix(param, n.Name)
}

func (n *DefaultNode) WithRunFunc(runFunc func(ctx context.Context, params map[string]any) (result map[string]any, err error)) *DefaultNode {
	n.runFunc = runFunc
	return n
}

func (n *DefaultNode) WithDetailFunc(detailFunc func(params map[string]any, result map[string]any) any) *DefaultNode {
	n.detailFunc = detailFunc
	return n
}

func (n *DefaultNode) Run(ctx context.Context) {
	if n.runFunc != nil {
		d := ctx.Value(DetailKey)

		params := n.loadParam(ctx)
		result, err := n.runFunc(ctx, params)
		if err != nil {
			ctx.Value(ChainKey).(*Chain).Stop()
			err = fmt.Errorf("node %s run failed: %w", n.Name, err)
			slog.Error(err.Error())
			if d != nil {
				d.(*Detail).Push(n.Name, map[string]any{"error": err.Error()})
			}
			return
		}
		if len(result) > 0 {
			n.setParam(ctx, result)
		}
		if n.detailFunc != nil {
			if d != nil {
				d.(*Detail).Push(n.Name, n.detailFunc(params, result))
			}
		}
	}

	ctx.Value(ChainKey).(*Chain).Next(ctx, n)
}
