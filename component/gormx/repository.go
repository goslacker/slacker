package gormx

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"reflect"
	"strings"

	"github.com/goslacker/slacker/core/database"
	"github.com/goslacker/slacker/core/tool/convert"
	"gorm.io/gorm"
)

func NewRepository[PO any, Entity any](db *gorm.DB, opts ...func(*Repository[PO, Entity])) *Repository[PO, Entity] {
	r := &Repository[PO, Entity]{
		DB:  db,
		ctx: context.Background(),
		M2E: database.DefaultM2E[PO, Entity],
		E2M: database.DefaultE2M[PO, Entity],
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type Repository[PO any, Entity any] struct {
	DB  *gorm.DB
	ctx context.Context
	M2E func(dst *Entity, src *PO) error
	E2M func(dst *PO, src *Entity) error
}

func (r *Repository[PO, Entity]) SetE2M(f any) {
	r.E2M = f.(func(dst *PO, src *Entity) error)
}

func (r *Repository[PO, Entity]) SetM2E(f any) {
	r.M2E = f.(func(dst *Entity, src *PO) error)
}

func (r *Repository[PO, Entity]) WithCtx(ctx context.Context) *Repository[PO, Entity] {
	tx := ctx.Value(database.TxKey)
	if tx != nil {
		return &Repository[PO, Entity]{
			DB:  tx.(*gorm.DB).WithContext(ctx),
			ctx: ctx,
			M2E: r.M2E,
			E2M: r.E2M,
		}
	}
	return &Repository[PO, Entity]{
		DB:  r.DB.WithContext(ctx),
		ctx: ctx,
		M2E: r.M2E,
		E2M: r.E2M,
	}
}

func (r *Repository[PO, Entity]) WithLock() *Repository[PO, Entity] {
	return &Repository[PO, Entity]{
		DB:  r.DB.Clauses(clause.Locking{Strength: "UPDATE"}),
		ctx: r.ctx,
		M2E: r.M2E,
		E2M: r.E2M,
	}
}

func (r *Repository[PO, Entity]) WithShareLock() *Repository[PO, Entity] {
	return &Repository[PO, Entity]{
		DB:  r.DB.Clauses(clause.Locking{Strength: "SHARE"}),
		ctx: r.ctx,
		M2E: r.M2E,
		E2M: r.E2M,
	}
}

func (r *Repository[PO, Entity]) Create(entities ...*Entity) (err error) {
	pos := make([]*PO, 0, len(entities))
	for _, item := range entities {
		po := new(PO)
		err = r.E2M(po, item)
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
		err = r.M2E(entities[index], item)
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
		err = r.E2M(po, x)
		if err != nil {
			return
		}
		err = r.DB.Updates(po).Error
		if err != nil {
			return
		}
		err = r.M2E(x, po)
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
	err = r.M2E(entity, po)
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
		err = r.M2E(&dst, &item)
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

func (r *Repository[PO, Entity]) PaginationByOffset(offset int, limit int, conditions ...any) (total int64, list []*Entity, err error) {
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
		err = r.M2E(&dst, &item)
		if err != nil {
			return
		}
		list = append(list, &dst)
	}

	return
}

func (r *Repository[PO, Entity]) Pagination(page int, size int, conditions ...any) (total int64, list []*Entity, err error) {
	if size == 0 {
		size = 15
	}
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * size
	return r.PaginationByOffset(offset, size, conditions...)
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

func (r *Repository[PO, Entity]) FirstOrCreate(entity *Entity, conditions ...any) (err error) {
	db, err := Apply(r.DB, conditions...)
	if err != nil {
		return
	}
	po := new(PO)
	err = r.E2M(po, entity)
	if err != nil {
		return
	}
	if len(conditions) == 0 {
		err = db.Where(po).FirstOrCreate(po).Error
	} else {
		err = db.FirstOrCreate(po).Error
	}
	if err != nil {
		return
	}
	err = r.M2E(entity, po)
	return
}

func (r *Repository[PO, Entity]) Save(entity *Entity) (err error) {
	po := new(PO)
	err = r.E2M(po, entity)
	if err != nil {
		return
	}
	err = r.DB.Save(po).Error
	if err != nil {
		return
	}
	err = r.M2E(entity, po)
	if err != nil {
		return
	}
	return
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
		case database.Limit:
			newDB = newDB.Limit(int(x))
		case database.Offset:
			newDB = newDB.Offset(int(x))
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
				switch strings.ToLower(c[1].(string)) {
				case "like":
					var value string
					value, err = convert.To[string](c[2])
					if err != nil {
						return
					}
					newDB = newDB.Where(fmt.Sprintf("%s LIKE (?)", quote(c[0].(string))), "%"+value+"%")
				case "not like":
					var value string
					value, err = convert.To[string](c[2])
					if err != nil {
						return
					}
					newDB = newDB.Where(fmt.Sprintf("%s NOT LIKE (?)", quote(c[0].(string))), "%"+value+"%")
				default:
					newDB = newDB.Where(fmt.Sprintf("%s %s (?)", quote(c[0].(string)), c[1]), c[2])
				}
			case 4:
				switch strings.ToLower(c[1].(string)) {
				case "between":
					newDB = newDB.Where(fmt.Sprintf("%s BETWEEN ? AND ?", quote(c[0].(string))), c[2], c[3])
				default:
					err = errors.New("condition params is too many")
					return
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
