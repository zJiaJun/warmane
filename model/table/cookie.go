package table

import "gorm.io/gorm"

const CookiesTableName = "cookies"

type Cookies struct {
	Host    string `gorm:"uniqueIndex:idx_host_name"`
	Name    string `gorm:"uniqueIndex:idx_host_name"`
	Cookies string
	gorm.Model
}

func (*Cookies) TableName() string {
	return CookiesTableName
}
