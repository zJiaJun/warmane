package model

type BasicCharacter struct {
	Realm             string `gorm:"uniqueIndex:idx_realm_name"`
	Name              string `gorm:"uniqueIndex:idx_realm_name"`
	Faction           string
	Race              string
	Gender            string
	Class             string
	Level             int
	AchievementPoints int
}
