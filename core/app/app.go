package app

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func NewApp() *App {
	return &App{
		components: make([]Component, 0, 20),
	}
}

type App struct {
	components []Component
	wg         sync.WaitGroup
}

func (a *App) RegisterComponent(components ...Component) {
	a.components = append(a.components, components...)
}

func (a *App) Init() (err error) {
	Fire(BeforeInit{})
	for _, module := range a.components {
		if m, ok := module.(Initable); ok {
			err = m.Init()
			if err != nil {
				return
			}
		}
	}
	Fire(AfterInit{})

	return
}

func (a *App) Boot() (err error) {
	Fire(BeforeBoot{})
	for _, module := range a.components {
		if m, ok := module.(Bootable); ok {
			err = m.Boot()
			if err != nil {
				return
			}
		}
	}
	Fire(AfterBoot{})

	return
}

func (a *App) Run() (n int, err error) {
	err = a.Init()
	if err != nil {
		return
	}

	err = a.Boot()
	if err != nil {
		return
	}

	Fire(BeforeRun{})
	defer Fire(AfterRun{})
	for _, m := range a.components {
		if module, ok := m.(Serviceable); ok {
			a.wg.Add(1)
			go func(start func()) {
				defer a.wg.Done()
				start()
			}(module.Start)
			n++
		}
	}

	return
}

func (a *App) Shutdown() {
	Fire(BeforeShutdown{})
	defer Fire(AfterShutdown{})
	for _, m := range a.components {
		if module, ok := m.(Serviceable); ok {
			go module.Stop()
		}
	}
}

func (a *App) RunAndWait() (err error) {
	n, err := a.Run()
	if err != nil {
		err = fmt.Errorf("run failed: %w", err)
		return
	}

	if n == 0 {
		err = errors.New("run failed: no service run")
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	for range signals {
		signal.Stop(signals)
		close(signals)
		a.Shutdown()
	}

	println("wait module stop...")
	a.wg.Wait()
	println("bye bye~")

	return
}
