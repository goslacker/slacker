package gormx

import (
	"context"
	"database/sql"

	"github.com/goslacker/slacker/core/database"
	"gorm.io/gorm"
)

type Holder interface {
	WithContext(ctx context.Context) *DB
	GetDB() *gorm.DB
}

func NewHolder(db *gorm.DB) Holder {
	return &DB{DB: db}
}

type DB struct {
	ctx context.Context
	*gorm.DB
}

func (h *DB) GetDB() *gorm.DB {
	return h.DB
}

func (h *DB) WithContext(ctx context.Context) *DB {
	// 从上下文中获取事务
	tx, ok := ctx.Value(database.TxKey).(*gorm.DB)
	if ok {
		return &DB{ctx: ctx, DB: tx.WithContext(ctx)}
	}
	return &DB{ctx: ctx, DB: h.DB.WithContext(ctx)}
}

func (h *DB) Transaction(f func(ctx context.Context) error, opts ...*sql.TxOptions) error {
	return h.DB.Transaction(func(tx *gorm.DB) error {
		ctx := context.WithValue(h.ctx, database.TxKey, tx)
		return f(ctx)
	}, opts...)
}

func (h *DB) Begin(opts ...*sql.TxOptions) (context.Context, error) {
	tx := h.DB.Begin(opts...)
	err := tx.Error
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(h.ctx, database.TxKey, tx)
	return ctx, tx.Error
}

func (h *DB) BeginCtx(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error) {
	return h.WithContext(ctx).Begin(opts...)
}

func (h *DB) TransactionCtx(ctx context.Context, f func(ctx context.Context) error, opts ...*sql.TxOptions) error {
	return h.WithContext(ctx).Transaction(f, opts...)
}

func (h *DB) Commit(ctx context.Context) error {
	return h.WithContext(ctx).DB.Commit().Error
}

func (h *DB) Rollback(ctx context.Context) error {
	return h.WithContext(ctx).DB.Rollback().Error
}
