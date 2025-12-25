package database

import (
	"context"
	"database/sql"
)

//go:generate mockgen -destination=tx_manager.go -package=databasex . TxManager

type TxManager interface {
	Begin(opts ...*sql.TxOptions) (newCtx context.Context, err error)
	BeginCtx(ctx context.Context, opts ...*sql.TxOptions) (newCtx context.Context, err error)
	Commit(ctx context.Context) (err error)
	Rollback(ctx context.Context) (err error)
	TransactionCtx(ctx context.Context, f func(ctx context.Context) error, opts ...*sql.TxOptions) (err error)
	Transaction(f func(ctx context.Context) error, opts ...*sql.TxOptions) (err error)
}
