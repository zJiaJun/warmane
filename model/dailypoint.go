package model

import (
	"gorm.io/gorm"
	"time"
)

const DailyPointTableName = "daily_point_info"

type DailyPoint struct {
	Account   *Account `gorm:"embedded"`
	PointDate time.Time
	gorm.Model
}

func (*DailyPoint) TableName() string {
	return DailyPointTableName
}
