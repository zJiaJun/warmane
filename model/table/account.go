package table

import "gorm.io/gorm"

const accountTableName = "account"

type Account struct {
	Host        string `gorm:"uniqueIndex:idx_account"`
	AccountName string `gorm:"uniqueIndex:idx_account"`
	Password    string
	Cookies     string
	Status      string `gorm:"not null"`
	gorm.Model
}

func (*Account) TableName() string {
	return accountTableName
}
