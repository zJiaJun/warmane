package table

import "gorm.io/gorm"

const accountDetailsTableName = "account_details"

type AccountDetails struct {
	AccountId     uint   `gorm:"uniqueIndex:idx_account_details"`
	AccountName   string `gorm:"uniqueIndex:idx_account_details"`
	Coins         string
	Points        string
	Email         string
	Status        string
	DonationRank  string
	ActivityRank  string
	CommunityRank string
	JoinDate      string
	LastSeen      string
	gorm.Model
}

func (*AccountDetails) TableName() string {
	return accountDetailsTableName
}
