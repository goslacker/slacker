package database

import "context"

type Repository[Entity any] interface {
	WithCtx(context.Context) Repository[Entity]
	Create(...*Entity) error
	Update(entityOrMap any, conditions ...any) error
	First(conditions ...any) (*Entity, error)
	List(conditions ...any) ([]*Entity, error)
	Delete(conditions ...any) error
	// PaginationByOffset 通过偏移分页查询
	PaginationByOffset(offset int, limit int, conditions ...any) (total int64, list []*Entity, err error)
	// Pagination 通过页数分页查询
	Pagination(page, size int, conditions ...any) (total int64, list []*Entity, err error)
	Transaction(f func(ctx context.Context) error) (err error)
	Begin() (ctx context.Context)
	Commit(ctx context.Context) (err error)
	Rollback(ctx context.Context) (err error)
	SetE2M(f any)
	SetM2E(f any)
}

type Order []string

type Condition []any

type Limit int

type Offset int

type Migrator interface {
	RegisterMigrates(...any) error
	Migrate() error
	Up(stepNum int) error
	Down(stepNum int) error
}
