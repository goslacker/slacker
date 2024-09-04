package gormx

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/database"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewComponent() *Component {
	return &Component{}
}

type Component struct {
	app.Component
}

func (c *Component) Init() (err error) {
	conf := viper.Sub("database")
	dsn := database.DSN(conf.GetString("dsn"))
	db, err := gorm.Open(mysql.Open(dsn.RemoveSchema()), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})
	if err != nil {
		return
	}

	app.Bind[*gorm.DB](db)

	sqlDb, _ := db.DB()
	app.Bind[*sql.DB](sqlDb)
	return
}
