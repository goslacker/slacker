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
	conf := viper.Sub("gormx")
	mysqlLogger := conf.GetStringMap("logger")
	// 默认配置
	defaultLogger := map[string]interface{}{
		"slow_threshold":                200,
		"log_level":                     4,
		"ignore_record_not_found_error": true,
		"colorful":                      true,
		"parameterized_queries":         false,
	}

	// 检查是否有缺失配置
	for key, value := range defaultLogger {
		if _, exists := mysqlLogger[key]; !exists {
			mysqlLogger[key] = value
		}
	}

	slowThreshold := mysqlLogger["slow_threshold"].(int)
	logLevel := mysqlLogger["log_level"].(int)
	ignoreRecordNotFoundError := mysqlLogger["ignore_record_not_found_error"].(bool)
	colorful := mysqlLogger["colorful"].(bool)
	parameterizedQueries := mysqlLogger["parameterized_queries"].(bool)
	dsn := database.DSN(conf.GetString("dsn"))

	db, err := gorm.Open(mysql.Open(dsn.RemoveSchema()), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             time.Duration(slowThreshold) * time.Millisecond,
			LogLevel:                  logger.LogLevel(logLevel),
			IgnoreRecordNotFoundError: ignoreRecordNotFoundError,
			Colorful:                  colorful,
			ParameterizedQueries:      parameterizedQueries,
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
