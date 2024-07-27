package client

import "fmt"

type Initializer func(apiKey string) AIClient

var initializers = map[string]Initializer{} //map[model-name]Initalizer

func Register(model string, initializer Initializer) {
	initializers[model] = initializer
}

func New(model string, apiKey string) AIClient {
	i, ok := initializers[model]
	if !ok {
		panic(fmt.Errorf("client of model <%s> not found", model))
	}

	return i(apiKey)
}
