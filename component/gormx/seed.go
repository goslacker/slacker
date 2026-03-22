package gormx

import (
	"github.com/goslacker/slacker/core/app"
	"gorm.io/gorm"
)

func NewSeedComponent() *SeedComponent {
	return &SeedComponent{}
}

type SeedComponent struct {
	app.Component
}

func (s *SeedComponent) Init() (err error) {
	err = app.Bind[*SeedManager](func() *SeedManager {
		return &SeedManager{}
	})
	if err != nil {
		return
	}
	return
}

func (s *SeedComponent) Boot() (err error) {
	manager, err := app.Resolve[*SeedManager]()
	if err != nil {
		return
	}
	db, err := app.Resolve[*gorm.DB]()
	if err != nil {
		return
	}
	err = manager.Seed(db)
	if err != nil {
		return
	}
	return
}

func (s *SeedComponent) Seed(db *gorm.DB) error {
	return s.Seed(db)
}

type SeedManager struct {
	Seeds []func(db *gorm.DB) error
}

func (s *SeedManager) RegisterSeed(seeds ...func(db *gorm.DB) error) {
	s.Seeds = append(s.Seeds, seeds...)
}

func (s *SeedManager) Seed(db *gorm.DB) error {
	for _, seed := range s.Seeds {
		if err := seed(db); err != nil {
			return err
		}
	}
	return nil
}
