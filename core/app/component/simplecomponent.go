package component

import "github.com/goslacker/slacker/core/app"

type simpleComponent struct {
	app.Component
	init func() error
	boot func() error
}

func (sc *simpleComponent) Init() error {
	if sc.init == nil {
		return nil
	}
	return sc.init()
}

func (sc *simpleComponent) Boot() error {
	if sc.boot == nil {
		return nil
	}
	return sc.boot()
}
