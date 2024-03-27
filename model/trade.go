package model

import "gorm.io/gorm"

const TradeInfoTableName = "trade_info"

type TradeInfo struct {
	Name      string `gorm:"uniqueIndex"`
	ArmoryUrl string
	Coins     int
	CharDesc  string
	gorm.Model
}

func (*TradeInfo) TableName() string {
	return TradeInfoTableName
}
