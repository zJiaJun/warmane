package engine

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/storage"
	"github.com/zJiajun/warmane/constant"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/model/table"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (e *Engine) ListAccount() ([]*table.Account, error) {
	var account []*table.Account
	if result := e.db.Find(&account); result.Error != nil {
		return nil, result.Error
	}
	return account, nil
}

func (e *Engine) ListOnlineAccount() ([]*table.Account, error) {
	var account []*table.Account
	if result := e.db.Where("status = ?", constant.ONLINE_STATUS).Find(&account); result.Error != nil {
		return nil, result.Error
	}
	return account, nil
}

func (e *Engine) CreateAccount(account *table.Account) (int64, error) {
	if result := e.db.Create(account); result.Error != nil {
		return 0, result.Error
	} else {
		return result.RowsAffected, nil
	}
}

func (e *Engine) UpdateAccount(account *table.Account) (int64, error) {
	if result := e.db.Model(&table.Account{}).Where("id = ?", account.ID).Updates(account); result.Error != nil {
		return 0, result.Error
	} else {
		return result.RowsAffected, nil
	}
}

func (e *Engine) DeleteAccount(accountId int64) (int64, error) {
	if result := e.db.Delete(&table.Account{}, accountId); result.Error != nil {
		return 0, result.Error
	} else {
		return result.RowsAffected, nil
	}
}

var onceMap = make(map[int64]sync.Once)

func (e *Engine) CheckAccount(accountId int64) (bool, error) {
	var account *table.Account
	e.db.First(&account, accountId)
	if err := e.validateCookiesKey(account.Cookies); err != nil {
		return false, err
	}
	accountDetails := &table.AccountDetails{
		AccountId:   account.ID,
		AccountName: account.AccountName,
	}
	if err := e.validAndFetchAccountInfo(account.AccountName, accountDetails); err != nil {
		return false, err
	}
	if err := e.updateAccountInfo(account, accountDetails); err != nil {
		return false, err
	}
	if _, ok := onceMap[accountId]; !ok {
		e.keepingAccountOnline(accountId)
	}
	return true, nil
}

func (e *Engine) keepingAccountOnline(accountId int64) {
	once, ok := onceMap[accountId]
	if !ok {
		onceMap[accountId] = sync.Once{}
	}
	once.Do(func() {
		t := time.NewTicker(time.Duration(rand.IntN(12)) * time.Minute)
		for {
			select {
			case <-t.C:
				if _, err := e.CheckAccount(accountId); err != nil {
					logger.Errorf("keeping account [%d] online error, reasion: %v", accountId, err)
				}
			}
		}
	})
}

func (e *Engine) validateCookiesKey(cookies string) error {
	if cookies == "" {
		return errors.New("Cookies not exist or account not exist")
	}
	httpCookies := storage.UnstringifyCookies(cookies)
	for _, v := range constant.CookieKeys {
		if !storage.ContainsCookie(httpCookies, v) {
			return errors.New("Cookies is incorrect, key is missing: " + v)
		}
	}
	return nil
}

func (e *Engine) validAndFetchAccountInfo(accountName string, accountDetails *table.AccountDetails) error {
	isLogin, _ := false, false
	c := e.getScraper(accountName).CloneCollector()

	c.OnResponse(func(response *colly.Response) {
		logger.Info((string)(response.Body))
	})
	c.OnHTML("div.content-inner.left > table > tbody > tr:nth-child(2) > td", func(element *colly.HTMLElement) {
		isLogin = strings.Contains(element.Text, accountName)
	})
	c.OnHTML("form[id='frmAuthenticate']", func(element *colly.HTMLElement) {
		//isAuth = true
	})

	c.OnHTML(".myCoins", func(element *colly.HTMLElement) {
		accountDetails.Coins = element.Text
	})
	c.OnHTML(".myPoints", func(element *colly.HTMLElement) {
		accountDetails.Points = element.Text
	})
	c.OnHTML("div.content-inner.left > table > tbody > tr:nth-child(6) > td > a", func(element *colly.HTMLElement) {
		accountDetails.Email = strings.TrimSpace(decodeEmail(element.Attr("data-cfemail")))
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(2) > td", func(element *colly.HTMLElement) {
		accountDetails.Status = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(3) > td", func(element *colly.HTMLElement) {
		accountDetails.DonationRank = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(4) > td", func(element *colly.HTMLElement) {
		accountDetails.ActivityRank = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(5) > td", func(element *colly.HTMLElement) {
		accountDetails.CommunityRank = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(7) > td", func(element *colly.HTMLElement) {
		accountDetails.JoinDate = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(8) > td", func(element *colly.HTMLElement) {
		accountDetails.LastSeen = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	if err := c.Visit(constant.AccountUrl); err != nil {
		return err
	}
	if !isLogin {
		return errors.New("Account can not login")
	}
	logger.Infof("Account [%s] login success, fetch account info", accountName)
	return nil
}

func decodeEmail(encodedEmail string) string {
	if encodedEmail == "" {
		return "[email protected]"
	}
	r, _ := strconv.ParseInt(encodedEmail[0:2], 16, 0)
	n := 2
	decoded := ""
	for n < len(encodedEmail) {
		part, _ := strconv.ParseInt(encodedEmail[n:n+2], 16, 0)
		decoded += "%" + fmt.Sprintf("%0.2x", int(part)^int(r))
		n += 2
	}
	unquoted, _ := url.QueryUnescape(decoded)
	return unquoted
}

func (e *Engine) updateAccountInfo(account *table.Account, accountDetails *table.AccountDetails) error {
	result := e.db.Model(&table.Account{}).Where("id = ?", account.ID).Where("status = ?", constant.OFFLINE_STATUS).Update("status", constant.ONLINE_STATUS)
	if result.Error != nil {
		return result.Error
	}
	if err := e.addOrUpdateAccountDetails(accountDetails); err != nil {
		return err
	}
	return nil
}
