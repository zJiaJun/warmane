package engine

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/captcha"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/scraper"
	"os"
	"sync"
)

type BodyMsg struct {
	Messages struct {
		Success []string `json:"success"`
		Error   []string `json:"error"`
	}
	Points []float64 `json:"points"`
}

const (
	/*
			{"messages":{"errors":["Incorrect account name or password."]}}
			{"messages":{"errors":["You have already collected your points today."]}}
		 	{"messages":{"errors":["You have not logged in-game today."]}}
			{"messages":{"success":["Daily points collected."]},"points":[10.4]}
	*/
	loginSuccessBody = "{\"redirect\":[\"\\/account\"]}"
)

type Engine struct {
	config    *config.Config
	scrapers  map[string]*scraper.Scraper
	wg        sync.WaitGroup
	csrfToken string
}

func New() *Engine {
	conf, err := config.LoadConf()
	if err != nil {
		panic(err)
	}
	return &Engine{
		config:   conf,
		scrapers: make(map[string]*scraper.Scraper, len(conf.Accounts)),
	}
}

func (e *Engine) RunDailyPoints() {
	glog.Info("开始运行自动签到功能")
	defer glog.Flush()
	count := len(e.config.Accounts)
	glog.Infof("加载配置文件[config.yml]成功, 需要签到的账号数量是[%d]", count)
	e.wg.Add(count)
	glog.Infof("开始goroutine并发处理")
	for _, account := range e.config.Accounts {
		e.setScraper(account)
		go e.collect(account)
	}
	e.wg.Wait()
}

func (e *Engine) setScraper(account config.Account) {
	e.scrapers[account.Username] = scraper.NewScraper(account.Username)
}

func (e *Engine) scraper(account config.Account) *scraper.Scraper {
	return e.scrapers[account.Username]
}

func (e *Engine) collect(account config.Account) {
	defer e.wg.Done()
	if err := e.login(account); err != nil {
		glog.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.collectPoints(account); err != nil {
		glog.Errorf("账号[%s]自动收集签到点错误, 原因: %v", account.Username, err)
		return
	}
	/*
		if err := e.trade(account); err != nil {
			glog.Errorf("账号[%s]查询商场数据错误, 原因: %v", account.Username, err)
			return
		}
	*/
	/*
		if err := e.logout(account); err != nil {
			glog.Errorf("账号[%s]退出错误, 原因: %v", account.Username, err)
			return
		}
	*/
}

func (e *Engine) login(account config.Account) error {
	if e.existCookiesFile(account) {
		glog.Infof("存在[%s]cookies文件,跳过登录", account.Username)
		return nil
	}
	c := e.scraper(account).CloneCollector()
	var err error
	e.csrfToken, err = getCsrfToken(c)
	if err != nil {
		return err
	}

	capt := captcha.NewCaptcha(e.config.CaptchaApiKey, e.config.WarmaneSiteKey, config.LoginUrl)
	code, err := capt.HandleCaptcha()
	if err != nil {
		return err
	}
	loginData := map[string]string{
		"return":               "",
		"userID":               account.Username,
		"userPW":               account.Password,
		"g-recaptcha-response": code,
		"userRM":               "on",
	}
	e.scraper(account).SetRequestHeaders(c, e.csrfToken)
	e.scraper(account).DecodeResponse(c)
	var bodyMsg BodyMsg
	c.OnResponse(func(response *colly.Response) {
		bodyText := string(response.Body)
		if bodyText == loginSuccessBody {
			glog.Infof("账号[%s]登录成功", account.Username)
		} else {
			err := json.Unmarshal(response.Body, &bodyMsg)
			if err != nil {
				glog.Errorf("账号[%s]登陆解码Json错误, 返回内容: %s", account.Username, bodyText)
				return
			}
			if len(bodyMsg.Messages.Error) > 0 {
				errMsg := bodyMsg.Messages.Error[0]
				glog.Infof("账号[%s]登录失败, %s", account.Username, errMsg)
			} else {
				glog.Infof("账号[%s]登录失败, %s", account.Username, bodyText)
			}
		}
	})
	err = c.Post(config.LoginUrl, loginData)
	return err
}

func (e *Engine) existCookiesFile(account config.Account) bool {
	cookiesFile := fmt.Sprintf("www.warmane.com.%s.cookies", account.Username)
	_, err := os.Stat(cookiesFile)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (e *Engine) collectPoints(account config.Account) error {
	beforeCoins, beforePoints, err := e.getInfo(account)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[前]的 coins: [%s], points: [%s]",
		account.Username, beforeCoins, beforePoints)

	c := e.scraper(account).CloneCollector()
	e.scraper(account).SetRequestHeaders(c, e.csrfToken)
	e.scraper(account).DecodeResponse(c)
	var bodyMsg BodyMsg
	c.OnResponse(func(response *colly.Response) {
		bodyText := string(response.Body)
		err := json.Unmarshal(response.Body, &bodyMsg)
		if err != nil {
			glog.Errorf("账号[%s]收集签到解码Json错误, 返回内容: %s", account.Username, bodyText)
			return
		}
		if len(bodyMsg.Messages.Success) > 0 && len(bodyMsg.Points) > 0 {
			successMsg := bodyMsg.Messages.Success[0]
			points := bodyMsg.Points[0]
			glog.Infof("账号[%s]自动收集签到点成功, 返回内容: %s, 签到点: %f", account.Username, successMsg, points)
		} else if len(bodyMsg.Messages.Error) > 0 {
			errorMsg := bodyMsg.Messages.Error[0]
			glog.Infof("账号[%s]自动收集签到点失败, 返回内容: %s", account.Username, errorMsg)
		} else {
			glog.Infof("账号[%s]自动收集签到点失败, 返回内容: %s", account.Username, bodyText)
		}
	})
	collectPointsData := map[string]string{"collectpoints": "true"}
	err = c.Post(config.AccountUrl, collectPointsData)
	if err != nil {
		return err
	}
	afterCoins, afterPoints, err := e.getInfo(account)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[后]的 coins: [%s], points: [%s]",
		account.Username, afterCoins, afterPoints)
	return err
}

func (e *Engine) trade(account config.Account) error {
	c := e.scraper(account).CloneCollector()
	c.OnResponse(func(response *colly.Response) {
		respBody := response.Body
		glog.Infof(string(respBody))
	})
	err := c.Visit(config.TradeUrl)
	return err
}

func (e *Engine) logout(account config.Account) error {
	c := e.scraper(account).CloneCollector()
	err := c.Visit(config.LogoutUrl)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]退出成功", account.Username)
	return err
}

func (e *Engine) getInfo(account config.Account) (coins string, points string, err error) {
	c := e.scraper(account).CloneCollector()
	e.scraper(account).SetRequestHeaders(c, e.csrfToken)
	e.scraper(account).DecodeResponse(c)
	c.OnHTML(config.CoinsSelector, func(element *colly.HTMLElement) {
		coins = element.Text
	})
	c.OnHTML(config.PointsSelector, func(element *colly.HTMLElement) {
		points = element.Text
	})
	err = c.Visit(config.AccountUrl)
	return
}

func getCsrfToken(c *colly.Collector) (csrfToken string, err error) {
	c.OnHTML(config.CsrfTokenSelector, func(element *colly.HTMLElement) {
		csrfToken = element.Attr("content")
	})
	err = c.Visit(config.LoginUrl)
	glog.Infof("查询获取warmane网站的csrfToken成功: %s", csrfToken)
	return
}
