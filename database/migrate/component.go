package migrate

import (
	"log/slog"

	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/database"
)

func NewComponent() *Component {
	m := &Component{}

	return m
}

type Component struct {
	app.Component
}

func (m Component) Init() (err error) {
	err = app.Bind[database.Migrator](NewDefaultMigrator)
	if err != nil {
		return
	}

	app.RegisterListener(func(event app.BeforeBoot) {
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
