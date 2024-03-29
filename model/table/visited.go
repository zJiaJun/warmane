package table

import "gorm.io/gorm"

const VisitedTableName = "visited"

type Visited struct {
	RequestID int `gorm:"index"`
	Visited   int
	gorm.Model
}

func (*Visited) TableName() string {
	return VisitedTableName
}
