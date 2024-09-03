package migrate

import (
	"log/slog"

	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/database"
)

func NewComponent() *Module {
	m := &Module{}

	return m
}

type Module struct {
	app.Component
}

func (m Module) Init() (err error) {
	err = app.Bind[database.Migrator](NewDefaultMigrator)
	if err != nil {
		return
	}

	app.RegisterListener(func(event app.AfterInit) {
		err = app.Invoke(func(m database.Migrator) (err error) {
			err = m.Migrate()
			if err != nil {
				return
			}
			return
		})
		if err != nil {
			slog.Error("migrate failed", "err", err)
		}
	})

	return
}
