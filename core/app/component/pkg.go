package component

import (
	"github.com/goslacker/slacker/core/app"
)

func Init(init func() error) *simpleComponent {
	return &simpleComponent{
		init: init,
	}
}

func Boot(boot func() error) *simpleComponent {
	return &simpleComponent{
		boot: boot,
	}
}

func SimpleComponent(init, boot func() error) *simpleComponent {
	return &simpleComponent{
		init: init,
		boot: boot,
	}
}

func AfterInit(afterInit func()) *simpleComponent {
	return &simpleComponent{
		init: func() error {
			app.RegisterListener(func(event app.AfterInit) {
				afterInit()
			})
			return nil
		},
	}
}
