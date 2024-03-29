package table

import (
	"gitub.com/zJiajun/warmane/model"
	"gorm.io/gorm"
	"time"
)

const DailyPointTableName = "daily_point_info"

type DailyPoint struct {
	Account   *model.Account `gorm:"embedded"`
	PointDate time.Time
	gorm.Model
}

func (*DailyPoint) TableName() string {
	return DailyPointTableName
}
