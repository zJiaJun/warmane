package model

type Character struct {
	BasicCharacter *BasicCharacter `gorm:"embedded"`
	Guild          string
	Professions    string
}
