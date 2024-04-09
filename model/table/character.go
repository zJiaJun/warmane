package table

import "github.com/zJiajun/warmane/model"

type Character struct {
	BasicCharacter *model.BasicCharacter `gorm:"embedded"`
}
