package gormx

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"testing"
)

func TestNormal(t *testing.T) {
	db, err := gorm.Open(mysql.Open("root:toor@tcp(127.0.0.1:3306)/analysis?charset=utf8mb4&parseTime=True&loc=Local"))
	require.NoError(t, err)

	fmt.Printf("%+v\n", db.Statement)
	db = db.Where("a = ?", 123)
	fmt.Printf("%+v\n", db.Statement)
	tx := db.Begin()
	fmt.Printf("%+v\n", tx.Statement)
	ttx := tx.Session(&gorm.Session{})
	fmt.Printf("%+v\n", ttx.Statement)

	ttx.Statement = &gorm.Statement{
		DB:        ttx.Statement.DB,
		ConnPool:  db.Statement.ConnPool,
		Context:   db.Statement.Context,
		Clauses:   map[string]clause.Clause{},
		Vars:      make([]interface{}, 0, 8),
		SkipHooks: db.Statement.SkipHooks,
	}
	fmt.Printf("%+v\n", ttx.Statement)
}
