package engine

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/logger"
	"gitub.com/zJiajun/warmane/model"
	"strings"
)

func (e *Engine) RunDailyPoints() {
	logger.Info("开始运行自动签到功能")
	count := len(e.config.Accounts)
	logger.Infof("加载配置文件[config.yml]成功, 需要签到的账号数量是[%d]", count)
	e.wg.Add(count)
	logger.Infof("开始goroutine并发处理")
	for _, v := range e.config.Accounts {
		go e.collectPoints(v)
	}
	e.wg.Wait()
}

func (e *Engine) collectPoints(account config.Account) {
	defer e.wg.Done()
	if err := e.login(account); err != nil {
		logger.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.collect(account); err != nil {
		logger.Errorf("账号[%s]自动收集签到点错误, 原因: %v", account.Username, err)
		return
	}
}

func (e *Engine) collect(account config.Account) error {
	name := account.Username
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
			logger.Errorf("账号[%s]收集签到解码Json错误, 返回内容: %s", name, bodyText)
			return
		}
		if len(bodyMsg.Messages.Success) > 0 && len(bodyMsg.Points) > 0 {
			successMsg := bodyMsg.Messages.Success[0]
			points := bodyMsg.Points[0]
			logger.Infof("账号[%s]自动收集签到点成功, 返回内容: %s, 签到点: %f", name, successMsg, points)
		} else if len(bodyMsg.Messages.Error) > 0 {
			errorMsg := bodyMsg.Messages.Error[0]
			logger.Infof("账号[%s]自动收集签到点失败, 返回内容: %s", name, errorMsg)
		} else {
			logger.Infof("账号[%s]自动收集签到点失败, 返回内容: %s", name, bodyText)
		}
	})
	collectPointsData := map[string]string{"collectpoints": "true"}
	if err := c.Post(constant.AccountUrl, collectPointsData); err != nil {
		return err
	}
	acc, err := e.getAccountInfo(account)
	if err != nil {
		return err
	}
	logger.Infof("账号[%s]收集签到点[后]的信息 %s", name, acc)
	return err
}

func (e *Engine) getAccountInfo(account config.Account) (*model.Account, error) {
	acc := &model.Account{}
	name := account.Username
	acc.Name = name
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	c.OnHTML(".myCoins", func(element *colly.HTMLElement) {
		acc.Coins = element.Text
	})
	c.OnHTML(".myPoints", func(element *colly.HTMLElement) {
		acc.Points = element.Text
	})
	c.OnHTML("div.content-inner.left > table > tbody > tr:nth-child(6) > td", func(element *colly.HTMLElement) {
		acc.Email = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(2) > td", func(element *colly.HTMLElement) {
		acc.Status = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(3) > td", func(element *colly.HTMLElement) {
		acc.DonationRank = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(4) > td", func(element *colly.HTMLElement) {
		acc.ActivityRank = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(5) > td", func(element *colly.HTMLElement) {
		acc.CommunityRank = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(7) > td", func(element *colly.HTMLElement) {
		acc.JoinDate = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	c.OnHTML("div.content-inner.right > table > tbody > tr:nth-child(8) > td", func(element *colly.HTMLElement) {
		acc.LastSeen = strings.TrimSpace(strings.Split(element.Text, ":")[1])
	})
	err := c.Visit(constant.AccountUrl)
	return acc, err
}
