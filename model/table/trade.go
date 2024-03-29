package table

import (
	"gitub.com/zJiajun/warmane/model"
	"gorm.io/gorm"
)

const TradeInfoTableName = "trade_info"

type TradeInfo struct {
	BasicCharacter    *model.BasicCharacter `gorm:"embedded"`
	ArmoryUrl         string
	Coins             int
	InventoryIncluded int
	CharDesc          string
	gorm.Model
}

func (*TradeInfo) TableName() string {
	return TradeInfoTableName
}
