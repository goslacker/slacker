package gormx

import (
	"context"
	"errors"

	"github.com/goslacker/slacker/core/tool"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("not found")

type Pagination struct {
	Page int
	Size int
}

func NewPagination(page int, size int) Pagination {
	return Pagination{
		Page: page,
		Size: size,
	}
}

func (p Pagination) Validate() bool {
	return p.Page > 0 || p.Size > 0
}

func (p Pagination) Offset() int {
	offset := (p.Page - 1) * p.Size
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (p Pagination) Limit() int {
	return p.Size
}

func (p Pagination) SetQuery(query *gorm.DB) *gorm.DB {
	return query.Offset(p.Offset()).Limit(p.Limit())
}

type PaginationCondition interface {
	Validate() bool
	SetQuery(query *gorm.DB) *gorm.DB
}

type HolderRepository[PO any, Entity any, Condition any] struct {
	DB             *DB
	parseCondition func(query *gorm.DB, condition Condition) *gorm.DB
}

func NewHolderRepository[PO any, Entity any, Condition any](db *DB, parseCondition func(query *gorm.DB, condition Condition) *gorm.DB) *HolderRepository[PO, Entity, Condition] {
	return &HolderRepository[PO, Entity, Condition]{
		DB:             db,
		parseCondition: parseCondition,
	}
}

func (h *HolderRepository[PO, Entity, Condition]) BuildQuery(ctx context.Context, condition Condition) *gorm.DB {
	query := h.DB.WithContext(ctx).GetDB()
	query = h.parseCondition(query, condition)
	return query
}

func (h *HolderRepository[PO, Entity, Condition]) Save(ctx context.Context, entities ...*Entity) error {
	if len(entities) == 0 {
		return nil
	}
	return tool.SimpleMapFuncBack(entities, func(dest []*PO) (err error) {
		return h.DB.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&dest).Error
	})
}

func (h *HolderRepository[PO, Entity, Condition]) SaveInBatches(ctx context.Context, batchSize int, entities ...*Entity) error {
	h = &HolderRepository[PO, Entity, Condition]{
		DB:             NewHolder(h.DB.GetDB().Session(&gorm.Session{CreateBatchSize: batchSize})),
		parseCondition: h.parseCondition,
	}
	return h.Save(ctx, entities...)
}

func (h *HolderRepository[PO, Entity, Condition]) Delete(ctx context.Context, condition Condition) error {
	query := h.BuildQuery(ctx, condition)
	return query.Delete(new(PO)).Error
}

func (h *HolderRepository[PO, Entity, Condition]) Find(ctx context.Context, condition Condition) (entity *Entity, err error) {
	query := h.BuildQuery(ctx, condition)
	var model PO
	err = query.First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Join(ErrNotFound, err)
		}
		return
	}
	err = tool.SimpleMap(&entity, model)
	return
}

func (h *HolderRepository[PO, Entity, Condition]) List(ctx context.Context, condition Condition) (entities []*Entity, err error) {
	query := h.BuildQuery(ctx, condition)
	var models []*PO
	err = query.Find(&models).Error
	if err != nil {
		return nil, err
	}
	err = tool.SimpleMap(&entities, models)
	return
}

func (h *HolderRepository[PO, Entity, Condition]) Count(ctx context.Context, condition Condition) (count int64, err error) {
	query := h.BuildQuery(ctx, condition)
	err = query.Model(new(PO)).Count(&count).Error
	return
}

func (h *HolderRepository[PO, Entity, Condition]) Pagination(ctx context.Context, condition Condition) (total int64, list []*Entity, err error) {
	pagination, ok := any(condition).(PaginationCondition)
	if !ok {
		err = errors.New("condition must be PaginationCondition")
		return
	}
	if !pagination.Validate() {
		err = errors.New("pagination is invalid")
		return
	}
	query := h.BuildQuery(ctx, condition)
	err = query.Model(new(PO)).Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	var models []*PO
	err = pagination.SetQuery(query).Find(&models).Error
	if err != nil {
		return
	}
	err = tool.SimpleMap(&list, models)
	if err != nil {
		return
	}
	return
}

func (h *HolderRepository[PO, Entity, Condition]) UpdatesByEntity(ctx context.Context, condition Condition, entity *Entity) error {
	return h.updates(ctx, condition, entity)
}

func (h *HolderRepository[PO, Entity, Condition]) UpdatesByMap(ctx context.Context, condition Condition, updates map[string]any) error {
	return h.updates(ctx, condition, updates)
}

func (h *HolderRepository[PO, Entity, Condition]) updates(ctx context.Context, condition Condition, entOrMap any) (err error) {
	query := h.BuildQuery(ctx, condition)
	switch x := entOrMap.(type) {
	case *Entity:
		err = tool.SimpleMapFuncBack(x, func(dest *PO) (err error) {
			return query.Updates(x).Error
		})
	case map[string]any:
		err = query.Model(new(PO)).Updates(x).Error
	default:
		err = errors.New("entOrMap must be *Entity or map[string]any")
	}
	return
}

func (h *HolderRepository[PO, Entity, Condition]) CreateInBatches(ctx context.Context, batchSize int, entities ...*Entity) error {
	return tool.SimpleMapFuncBack(entities, func(models []*PO) (err error) {
		return h.DB.WithContext(ctx).CreateInBatches(models, batchSize).Error
	})
}

func (h *HolderRepository[PO, Entity, Condition]) Batch(ctx context.Context, condition Condition, batchSize int, fn func(entities ...*Entity) error) error {
	var models []*PO
	query := h.BuildQuery(ctx, condition)
	return query.FindInBatches(&models, batchSize, func(tx *gorm.DB, batch int) error {
		var entities []*Entity
		err := tool.SimpleMap(&entities, models)
		if err != nil {
			return err
		}
		return fn(entities...)
	}).Error
}

func (h *HolderRepository[PO, Entity, Condition]) Exists(ctx context.Context, condition Condition) (result bool, err error) {
	query := h.BuildQuery(ctx, condition)
	var count int64
	err = query.Model(new(PO)).Count(&count).Error
	if err != nil {
		return
	}
	return count > 0, nil
}
