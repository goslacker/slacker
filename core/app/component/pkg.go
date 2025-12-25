package component

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
