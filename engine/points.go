package engine

import (
	"encoding/json"
	"errors"
	"github.com/gocolly/colly/v2"
	"github.com/zJiajun/warmane/constant"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/model/table"
	"gorm.io/gorm/clause"
	"strconv"
)

func (e *Engine) ListAccountDetails() ([]*table.AccountDetails, error) {
	var accountDetails []*table.AccountDetails
	if result := e.db.Find(&accountDetails); result.Error != nil {
		return nil, result.Error
	}
	return accountDetails, nil
}

func (e *Engine) addOrUpdateAccountDetails(accountDetails *table.AccountDetails) error {
	if result := e.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "account_name"}},
		UpdateAll: true,
	}).Create(accountDetails); result.Error != nil {
		return result.Error
	}
	return nil
}

func (e *Engine) updateAccountPoints(accountId int64, points string) error {
	result := e.db.Model(&table.AccountDetails{}).Where("account_id = ?", accountId).Update("points", points)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (e *Engine) CollectAccountPoints(accountId int64) error {
	points := ""
	var collectError error
	var account *table.Account
	e.db.First(&account, accountId)
	name := account.AccountName
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	var bodyMsg struct {
		Messages struct {
			Success []string `json:"success"`
			Error   []string `json:"error"`
		}
		Points []float64 `json:"points"`
	}
	c.OnResponse(func(response *colly.Response) {
		bodyText := string(response.Body)
		err := json.Unmarshal(response.Body, &bodyMsg)
		if err != nil {
			logger.Errorf("Account[%s] collect points, decode json error, response body: %s", name, bodyText)
			collectError = errors.New("decode json error")
			return
		}
		if len(bodyMsg.Messages.Success) > 0 && len(bodyMsg.Points) > 0 {
			successMsg := bodyMsg.Messages.Success[0]
			points = strconv.FormatFloat(bodyMsg.Points[0], 'f', -1, 64)
			logger.Infof("Account[%s] collect points, success: %s, points: %s", name, successMsg, points)
		} else if len(bodyMsg.Messages.Error) > 0 {
			errorMsg := bodyMsg.Messages.Error[0]
			logger.Errorf("Account[%s] collect points, error: %s", name, errorMsg)
			collectError = errors.New(errorMsg)
		} else {
			logger.Errorf("Account[%s] collect points, response body: %s", name, bodyText)
			collectError = errors.New("response error")
		}
	})
	collectPointsData := map[string]string{"collectpoints": "true"}
	if err := c.Post(constant.AccountUrl, collectPointsData); err != nil {
		return err
	}
	if collectError != nil {
		return collectError
	}
	if points != "" {
		if err := e.updateAccountPoints(accountId, points); err != nil {
			return err
		}
	}
	return nil
}
