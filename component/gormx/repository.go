package gormx

import (
	"context"
	"errors"
	"fmt"
	"github.com/goslacker/slacker/core/database"
	"github.com/goslacker/slacker/core/tool/convert"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

func NewRepository[PO any, Entity any](db *gorm.DB, opts ...func(database.Repository[Entity])) *Repository[PO, Entity] {
	r := &Repository[PO, Entity]{
		DB:  db,
		m2e: database.DefaultM2E[PO, Entity],
		e2m: database.DefaultE2M[PO, Entity],
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type Repository[PO any, Entity any] struct {
	DB  *gorm.DB
	ctx context.Context
	m2e func(dst *Entity, src *PO) error
	e2m func(dst *PO, src *Entity) error
}

func (r *Repository[PO, Entity]) SetE2M(f any) {
	r.e2m = f.(func(dst *PO, src *Entity) error)
}

func (r *Repository[PO, Entity]) SetM2E(f any) {
	r.m2e = f.(func(dst *Entity, src *PO) error)
}

func (r *Repository[PO, Entity]) WithCtx(ctx context.Context) database.Repository[Entity] {
	tx := ctx.Value(database.TxKey)
	if tx != nil {
		return &Repository[PO, Entity]{
			DB:  tx.(*gorm.DB).WithContext(ctx),
			ctx: ctx,
		}
	}
	return &Repository[PO, Entity]{
		DB:  r.DB.WithContext(ctx),
		ctx: ctx,
	}
}

func (r *Repository[PO, Entity]) Create(entities ...*Entity) (err error) {
	pos := make([]*PO, 0, len(entities))
	for _, item := range entities {
		po := new(PO)
		err = r.e2m(po, item)
		if err != nil {
			return
		}
		pos = append(pos, po)
	}

	err = r.DB.Create(&pos).Error
	if err != nil {
		return
	}

	for index, item := range pos {
		err = r.m2e(entities[index], item)
		if err != nil {
			return
		}
	}
	return
}

func (r *Repository[PO, Entity]) Update(entityOrMap any, conditions ...any) (err error) {
	switch x := entityOrMap.(type) {
	case *Entity:
		po := new(PO)
		err = r.e2m(po, x)
		if err != nil {
			return
		}
		err = r.DB.Updates(po).Error
		if err != nil {
			return
		}
		err = r.m2e(x, po)
		if err != nil {
			return
		}
	case map[string]any:
		query := r.DB
		if len(conditions) > 0 {
			query, err = Apply(query, conditions...)
			if err != nil {
				return
			}
		}
		err = query.Model(new(PO)).Updates(x).Error
	default:
		err = errors.New("only supported struct or map")
	}

	return
}

func (r *Repository[PO, Entity]) First(conditions ...any) (entity *Entity, err error) {
	db, err := Apply(r.DB, conditions...)
	if err != nil {
		return
	}

	po := new(PO)
	err = db.First(po).Error
	if err != nil {
		return
	}

	entity = new(Entity)
	err = r.m2e(entity, po)
	return
}

func (r *Repository[PO, Entity]) List(conditions ...any) (list []*Entity, err error) {
	db, err := Apply(r.DB, conditions...)
	if err != nil {
		return
	}

	var poList []PO
	err = db.Find(&poList).Error
	if err != nil {
		return
	}

	list = make([]*Entity, 0, len(poList))
	for _, item := range poList {
		var dst Entity
		err = r.m2e(&dst, &item)
		if err != nil {
			return
		}
		list = append(list, &dst)
	}

	return
}

func (r *Repository[PO, Entity]) Delete(conditions ...any) (err error) {
	db, err := Apply(r.DB, conditions...)
	if err != nil {
		return
	}

	err = db.Delete(new(PO)).Error
	return
}

func (r *Repository[PO, Entity]) Pagination(offset int, limit int, conditions ...any) (total int64, list []*Entity, err error) {
	db, err := Apply(r.DB, conditions...)
	if err != nil {
		return
	}
	err = db.Model(new(PO)).Count(&total).Error
	if err != nil {
		return
	}

	var poList []PO
	err = db.Offset(offset).Limit(limit).Find(&poList).Error
	if err != nil {
		return
	}

	list = make([]*Entity, 0, len(poList))
	for _, item := range poList {
		var dst Entity
		err = r.m2e(&dst, &item)
		if err != nil {
			return
		}
		list = append(list, &dst)
	}

	return
}

func (r *Repository[PO, Entity]) Transaction(f func(ctx context.Context) error) (err error) {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	err = r.DB.Transaction(func(tx *gorm.DB) (err error) {
		ctx = context.WithValue(ctx, database.TxKey, tx)
		return f(ctx)
	})

	return
}

func (r *Repository[PO, Entity]) Begin() (ctx context.Context) {
	ctx = r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, database.TxKey, r.DB.Begin())
}

func (r *Repository[PO, Entity]) Commit(ctx context.Context) (err error) {
	tx := ctx.Value(database.TxKey)
	if tx == nil {
		return
	}
	return tx.(*gorm.DB).Commit().Error
}

func (r *Repository[PO, Entity]) Rollback(ctx context.Context) (err error) {
	tx := ctx.Value(database.TxKey)
	if tx == nil {
		return
	}
	return tx.(*gorm.DB).Rollback().Error
}

func Apply(db *gorm.DB, conditions ...any) (newDB *gorm.DB, err error) {
	if len(conditions) == 0 {
		return db, nil
	}
	newDB = db
	for _, condition := range conditions {
		switch x := condition.(type) {
		case []any:
			newDB, err = applyCondition(newDB, x)
			if err != nil {
				return
			}
		case [][]any:
			xx := make([]database.Condition, 0, len(x))
			for _, item := range x {
				xx = append(xx, item)
			}
			newDB, err = applyCondition(newDB, xx...)
			if err != nil {
				return
			}
		case database.Condition:
			newDB, err = applyCondition(newDB, x)
			if err != nil {
				return
			}
		case []database.Condition:
			newDB, err = applyCondition(newDB, x...)
			if err != nil {
				return
			}
		case database.Order:
			for _, item := range x {
				newDB = newDB.Order(item)
			}
		default:
			err = fmt.Errorf("unsupported condition type: %T <%+v\n>", condition, condition)
			return
		}
	}
	return
}

func applyCondition(db *gorm.DB, conditions ...database.Condition) (newDB *gorm.DB, err error) {
	newDB = db
	for _, c := range conditions {
		if len(c) < 2 {
			return db, errors.New("condition require at least 2 params")
		}

		list := []string{
			" and ",
			" or ",
			"?",
			" not ",
			" between ",
			" like ",
			" is ",
		}
		if s, ok := c[0].(string); ok && contains(list, strings.ToLower(s)) {
			newDB = newDB.Where(s, c[1:]...)
		} else {
			switch len(c) {
			case 2:
				v := reflect.ValueOf(c[1])
				if v.Kind() == reflect.Slice {
					newDB = newDB.Where(fmt.Sprintf("%s IN ?", quote(c[0].(string))), c[1])
				} else {
					newDB = newDB.Where(fmt.Sprintf("%s = ?", quote(c[0].(string))), c[1])
				}
			case 3:
				switch c[1] {
				case "like":
					var value string
					value, err = convert.To[string](c[2])
					if err != nil {
						return
					}
					newDB = db.Where(fmt.Sprintf("%s LIKE (?)", quote(c[0].(string))), "%"+value+"%")
				default:
					newDB = newDB.Where(fmt.Sprintf("%s %s (?)", quote(c[0].(string)), c[1]), c[2])
				}
			default:
				err = errors.New("condition params is too many")
				return
			}
		}
	}

	return
}

func quote(field string) string {
	var arr []string
	for _, item := range strings.Split(field, ".") {
		arr = append(arr, "`"+item+"`")
	}
	return strings.Join(arr, ".")
}

func contains(target []string, str string) bool {
	for _, s := range target {
		if strings.Contains(str, s) {
			return true
		}
	}

	return false
}
