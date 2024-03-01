package engine

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/model"
)

func (e *Engine) RunDailyPoints() {
	glog.Info("开始运行自动签到功能")
	count := len(e.config.Accounts)
	glog.Infof("加载配置文件[config.yml]成功, 需要签到的账号数量是[%d]", count)
	e.wg.Add(count)
	glog.Infof("开始goroutine并发处理")
	for _, v := range e.config.Accounts {
		go e.collectPoints(v)
	}
	e.wg.Wait()
}

func (e *Engine) collectPoints(account config.Account) {
	defer e.wg.Done()
	if err := e.login(account); err != nil {
		glog.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.collect(account); err != nil {
		glog.Errorf("账号[%s]自动收集签到点错误, 原因: %v", account.Username, err)
		return
	}
	/*
		if err := e.logout(account); err != nil {
			glog.Errorf("账号[%s]退出错误, 原因: %v", account.Username, err)
			return
		}
	*/
}

func (e *Engine) collect(account config.Account) error {
	name := account.Username
	beforeCoins, beforePoints, err := e.getInfo(account)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[前]的 coins: [%s], points: [%s]", name, beforeCoins, beforePoints)

	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	var bodyMsg model.BodyMsg
	c.OnResponse(func(response *colly.Response) {
		bodyText := string(response.Body)
		err := json.Unmarshal(response.Body, &bodyMsg)
		if err != nil {
			glog.Errorf("账号[%s]收集签到解码Json错误, 返回内容: %s", name, bodyText)
			return
		}
		if len(bodyMsg.Messages.Success) > 0 && len(bodyMsg.Points) > 0 {
			successMsg := bodyMsg.Messages.Success[0]
			points := bodyMsg.Points[0]
			glog.Infof("账号[%s]自动收集签到点成功, 返回内容: %s, 签到点: %f", name, successMsg, points)
		} else if len(bodyMsg.Messages.Error) > 0 {
			errorMsg := bodyMsg.Messages.Error[0]
			glog.Infof("账号[%s]自动收集签到点失败, 返回内容: %s", name, errorMsg)
		} else {
			glog.Infof("账号[%s]自动收集签到点失败, 返回内容: %s", name, bodyText)
		}
	})
	collectPointsData := map[string]string{"collectpoints": "true"}
	err = c.Post(constant.AccountUrl, collectPointsData)
	if err != nil {
		return err
	}
	afterCoins, afterPoints, err := e.getInfo(account)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[后]的 coins: [%s], points: [%s]", name, afterCoins, afterPoints)
	return err
}

func (e *Engine) getInfo(account config.Account) (coins string, points string, err error) {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	c.OnHTML(constant.CoinsSelector, func(element *colly.HTMLElement) {
		coins = element.Text
	})
	c.OnHTML(constant.PointsSelector, func(element *colly.HTMLElement) {
		points = element.Text
	})
	err = c.Visit(constant.AccountUrl)
	return
}
