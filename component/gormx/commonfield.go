package gormx

import (
	"github.com/goslacker/slacker/core/app"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

var _ callbacks.BeforeCreateInterface = (*SnowflakeID)(nil)

type UnixTimestampMilli struct {
	CreatedAt int64 `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt int64 `json:"updated_at" gorm:"autoUpdateTime:milli"`
}

type UnixTimestamp struct {
	CreatedAt int64 `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64 `json:"updated_at" gorm:"autoUpdateTime"`
}

type CommonField struct {
	SnowflakeID
	UnixTimestamp
}

type SnowflakeID struct {
	ID uint64 `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
}

func (s *SnowflakeID) BeforeCreate(tx *gorm.DB) (err error) {
	snowflake, err := app.Resolve[*sonyflake.Sonyflake]()
	if err != nil {
		return
	}
	if s.ID == 0 {
		s.ID, err = snowflake.NextID()
	}
	return
}
