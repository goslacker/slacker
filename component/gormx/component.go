package gormx

import (
	"database/sql"
	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/database"
	"github.com/sony/sonyflake"
	"log"
	"os"
	"time"

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

	err = app.Bind[*gorm.DB](db)
	if err != nil {
		return
	}

	sqlDb, _ := db.DB()
	err = app.Bind[*sql.DB](sqlDb)
	if err != nil {
		return
	}

	err = app.Bind[*sonyflake.Sonyflake](func() *sonyflake.Sonyflake {
		return sonyflake.NewSonyflake(sonyflake.Settings{})
	})
	if err != nil {
		return
	}
	return
}
