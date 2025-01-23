package client

import (
	"fmt"
)

type Initializer func(apiKey string, options ...func(*NewOptions)) AIClient

var initializers = map[string]Initializer{} //map[model-name]Initalizer

func Register(model string, initializer Initializer) {
	initializers[model] = initializer
}

func New(model string, apiKey string, options ...func(*NewOptions)) AIClient {
	i, ok := initializers[model]
	if !ok {
		panic(fmt.Errorf("client of model <%s> not found", model))
	}

	opts := &NewOptions{}
	for _, o := range options {
		o(opts)
	}

	return i(apiKey)
}
