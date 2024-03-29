package table

import "gorm.io/gorm"

const CookiesTableName = "cookies"

type Cookies struct {
	Host    string `gorm:"index"`
	Name    string `gorm:"index"`
	Cookies string
	gorm.Model
}

func (*Cookies) TableName() string {
	return CookiesTableName
}
