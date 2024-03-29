package table

import "gitub.com/zJiajun/warmane/model"

type Character struct {
	BasicCharacter *model.BasicCharacter `gorm:"embedded"`
}
