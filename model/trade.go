package model

type TradeResp struct {
	Content []string `json:"content"`
}

type TradeInfo struct {
	Name      string    //角色名称
	ArmoryUrl string    //角色详情地址
	Coins     int       //售卖价格coins
	CharDesc  string    //角色概要
	Character Character //角色详情对象
}
