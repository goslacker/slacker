package migrate

import (
	"fmt"
	"log/slog"

	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/database"
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

	app.RegisterListener(func(event app.BeforeBoot) (err error) {
		err = app.Invoke(func(m database.Migrator) (err error) {
			err = m.Migrate()
			if err != nil {
				return
			}
			return
		})
		if err != nil {
			err = fmt.Errorf("migrate failed: %w", err)
			slog.Error(err.Error())
		}
		return
	})

	return
}
