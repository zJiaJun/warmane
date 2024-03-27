package model

import "encoding/json"

type Account struct {
	Name          string
	Coins         string
	Points        string
	Email         string
	Status        string
	DonationRank  string
	ActivityRank  string
	CommunityRank string
	JoinDate      string
	LastSeen      string
}

func (a Account) String() string {
	ab, err := json.Marshal(a)
	if err != nil {
		return "error"
	} else {
		return string(ab)
	}
}
