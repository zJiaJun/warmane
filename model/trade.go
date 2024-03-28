package model

import "gorm.io/gorm"

const TradeInfoTableName = "trade_info"

type TradeInfo struct {
	BasicCharacter
	ArmoryUrl         string
	Coins             int
	InventoryIncluded int
	CharDesc          string
	gorm.Model
}

func (*TradeInfo) TableName() string {
	return TradeInfoTableName
}
