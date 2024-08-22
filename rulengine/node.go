package ruleengine

import (
	"context"
)

type Node interface {
	GetID() string
	Run(ctx context.Context)
}

type DefaultNode struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	ParamMap map[string]string `json:"paramMap"`
	runFunc  func(ctx context.Context, params map[string]any) (result map[string]any, stop bool)
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

func (n *DefaultNode) SetRunFunc(runFunc func(ctx context.Context, params map[string]any) (result map[string]any, stop bool)) {
	n.runFunc = runFunc
}

func (n *DefaultNode) WithRunFunc(runFunc func(ctx context.Context, params map[string]any) (result map[string]any, stop bool)) *DefaultNode {
	n.runFunc = runFunc
	return n
}

func (n *DefaultNode) Run(ctx context.Context) {
	params := n.loadParam(ctx)
	result, stop := n.runFunc(ctx, params)
	if stop {
		ctx.Value(ChainKey).(*Chain).Stop()
		return
	}
	if len(result) > 0 {
		n.setParam(ctx, result)
	}
	ctx.Value(ChainKey).(*Chain).Next(ctx, n)
}
