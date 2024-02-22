package engine

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/internal/captcha"
	"gitub.com/zJiajun/warmane/internal/config"
	"gitub.com/zJiajun/warmane/internal/scraper"
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
	scraper   *scraper.Scraper
	wg        sync.WaitGroup
	csrfToken string
}

func New() *Engine {
	conf, err := config.LoadConf()
	if err != nil {
		panic(err)
	}
	return &Engine{
		config:  conf,
		scraper: scraper.NewScraper(),
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
		go e.loginAndCollect(account)
	}
	e.wg.Wait()
}

func (e *Engine) loginAndCollect(account config.Account) {
	defer e.wg.Done()
	if err := e.login(account); err != nil {
		glog.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.collectPoints(account); err != nil {
		glog.Errorf("账号[%s]自动收集签到点错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.logout(account); err != nil {
		glog.Errorf("账号[%s]退出错误, 原因: %v", account.Username, err)
		return
	}
}

func (e *Engine) login(account config.Account) error {
	c := e.scraper.CloneCollector()
	csrfToken, err := getCsrfToken(c)
	if err != nil {
		return err
	}
	e.csrfToken = csrfToken

	loginData := make(map[string]string, 4)
	loginData["return"] = ""
	loginData["userID"] = account.Username
	loginData["userPW"] = account.Password

	capt := captcha.NewCaptcha(e.config.CaptchaApiKey, e.config.WarmaneSiteKey, config.LoginUrl)
	code, err := capt.HandleCaptcha()
	if err != nil {
		return err
	}
	loginData["g-recaptcha-response"] = code
	e.scraper.SetRequestHeaders(c, e.csrfToken)
	e.scraper.DecodeResponse(c)
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

func (e *Engine) collectPoints(account config.Account) error {
	beforeCoins, beforePoints, err := e.getInfo()
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[前]的 coins: [%s], points: [%s]",
		account.Username, beforeCoins, beforePoints)

	c := e.scraper.CloneCollector()
	e.scraper.SetRequestHeaders(c, e.csrfToken)
	e.scraper.DecodeResponse(c)
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
	collectPointsData := map[string]string{
		"collectpoints": "true",
	}
	err = c.Post(config.AccountUrl, collectPointsData)
	if err != nil {
		return err
	}
	afterCoins, afterPoints, err := e.getInfo()
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[后]的 coins: [%s], points: [%s]",
		account.Username, afterCoins, afterPoints)
	return err
}

func (e *Engine) logout(account config.Account) error {
	c := e.scraper.CloneCollector()
	err := c.Visit(config.LogoutUrl)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]退出成功", account.Username)
	return err
}

func (e *Engine) getInfo() (coins string, points string, err error) {
	c := e.scraper.CloneCollector()
	e.scraper.SetRequestHeaders(c, e.csrfToken)
	e.scraper.DecodeResponse(c)
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
