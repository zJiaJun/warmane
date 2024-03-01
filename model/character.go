package model

type Character struct {
	Name              string //角色名称
	Faction           string //阵营
	Race              string //种族
	Class             string //职业
	Level             int    //等级
	AchievementPoints int    //成就点数
	ArenaPoints       int    //竞技场点数
	HonorPoints       int    //荣誉点数
}
