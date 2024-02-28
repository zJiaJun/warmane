package engine

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/storage"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/captcha"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/errors"
	"gitub.com/zJiajun/warmane/model"
	"gitub.com/zJiajun/warmane/scraper"
	"os"
	"sync"
)

type Engine struct {
	config    *config.Config
	scrapers  *scraper.Scrapers
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
		scrapers: scraper.New(conf.Accounts),
	}
}

func (e *Engine) getScraper(name string) *scraper.Scraper {
	return e.scrapers.Get(name)
}

func (e *Engine) RunDailyPoints() {
	glog.Info("开始运行自动签到功能")
	defer glog.Flush()
	count := len(e.config.Accounts)
	glog.Infof("加载配置文件[config.yml]成功, 需要签到的账号数量是[%d]", count)
	e.wg.Add(count)
	glog.Infof("开始goroutine并发处理")
	for _, v := range e.config.Accounts {
		go e.collect(v)
	}
	e.wg.Wait()
}

func (e *Engine) collect(account config.Account) {
	defer e.wg.Done()
	if token, err := e.initCsrfToken(account); err != nil {
		glog.Errorf("查询获取warmane网站的csrfToken错误, 原因: %v", err)
		return
	} else {
		e.csrfToken = token
		glog.Infof("查询获取warmane网站的csrfToken成功: %s", e.csrfToken)
	}

	if err := e.login(account); err != nil {
		glog.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	/*
		if err := e.collectPoints(account); err != nil {
			glog.Errorf("账号[%s]自动收集签到点错误, 原因: %v", account.Username, err)
			return
		}
	*/
	if err := e.trade(account); err != nil {
		glog.Errorf("账号[%s]查询商场数据错误, 原因: %v", account.Username, err)
		return
	}
	/*
		if err := e.logout(account); err != nil {
			glog.Errorf("账号[%s]退出错误, 原因: %v", account.Username, err)
			return
		}
	*/
}

func (e *Engine) login(account config.Account) error {
	name := account.Username
	if e.config.UseCookiesLogin {
		glog.Infof("配置项[useCookiesLogin]为true, 使用cookies文件登录")
		cookiesFile := constant.CookieFileName(name)
		if err := e.validateCookies(cookiesFile); err != nil {
			return err
		} else {
			glog.Infof("存在[%s]文件且验证通过,使用cookies文件登录", cookiesFile)
			return nil
		}
	} else {
		capt := captcha.NewCaptcha(e.config.CaptchaApiKey, e.config.WarmaneSiteKey, constant.LoginUrl)
		code, err := capt.HandleCaptcha()
		if err != nil {
			return err
		}
		loginData := map[string]string{
			"return":               "",
			"userID":               name,
			"userPW":               account.Password,
			"g-recaptcha-response": code,
			"userRM":               "on",
		}

		c := e.getScraper(name).CloneCollector()
		e.getScraper(name).SetRequestHeaders(c, e.csrfToken)
		e.getScraper(name).DecodeResponse(c)
		var bodyMsg model.BodyMsg
		c.OnResponse(func(response *colly.Response) {
			bodyText := string(response.Body)
			if bodyText == constant.LoginSuccessBody {
				glog.Infof("账号[%s]登录成功", name)
			} else {
				err := json.Unmarshal(response.Body, &bodyMsg)
				if err != nil {
					glog.Errorf("账号[%s]登陆解码Json错误, 返回内容: %s", name, bodyText)
					return
				}
				if len(bodyMsg.Messages.Error) > 0 {
					errMsg := bodyMsg.Messages.Error[0]
					glog.Infof("账号[%s]登录失败, %s", name, errMsg)
				} else {
					glog.Infof("账号[%s]登录失败, %s", name, bodyText)
				}
			}
		})
		err = c.Post(constant.LoginUrl, loginData)
		return err
	}
}

func (e *Engine) validateCookies(cookiesFile string) error {
	_, err := os.Stat(cookiesFile)
	if os.IsNotExist(err) {
		glog.Errorf("不存在[%s]文件", cookiesFile)
		return errors.ErrCookieNotFound
	}
	file, err := os.ReadFile(cookiesFile)
	if err != nil {
		return err
	}
	cookies := storage.UnstringifyCookies(string(file))
	for _, v := range constant.CookieKeys {
		if !storage.ContainsCookie(cookies, v) {
			glog.Infof("存在[%s]文件,不匹配cookieKey[%s]", cookiesFile, v)
			//_ = os.Remove(cookiesFile)
			return errors.ErrCookieNonMatchKey
		}
	}
	return nil
}

func (e *Engine) collectPoints(account config.Account) error {
	name := account.Username
	beforeCoins, beforePoints, err := e.getInfo(account)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]收集签到点[前]的 coins: [%s], points: [%s]", name, beforeCoins, beforePoints)

	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c, e.csrfToken)
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

func (e *Engine) trade(account config.Account) error {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c, e.csrfToken)
	e.getScraper(name).DecodeResponse(c)
	c.OnResponse(func(response *colly.Response) {
		respBody := response.Body
		glog.Infof(string(respBody))
	})
	searchTradeData := map[string]string{
		"update":         "page",
		"timeout":        "false",
		"hovering":       "false",
		"tradehandler":   "",
		"service":        "charactertrade",
		"currency":       "coins",
		"realm":          "7",
		"character":      "",
		"currentmenu":    "-1",
		"currentsubmenu": "-1",
		"class":          "-1",
		"purchasetype":   "0",
		"purchasevalue":  "0",
		"page":           "0",
		"tradetab":       "",
		"selltab":        "",
		"method":         "load",
		"do":             "search",
	}
	err := c.Post(constant.TradeUrl, searchTradeData)
	return err
}

func (e *Engine) logout(account config.Account) error {
	c := e.getScraper(account.Username).CloneCollector()
	err := c.Visit(constant.LogoutUrl)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]退出成功", account.Username)
	return err
}

func (e *Engine) getInfo(account config.Account) (coins string, points string, err error) {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c, e.csrfToken)
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

func (e *Engine) initCsrfToken(account config.Account) (csrfToken string, err error) {
	c := e.getScraper(account.Username).CloneCollector()
	c.OnHTML(constant.CsrfTokenSelector, func(element *colly.HTMLElement) {
		csrfToken = element.Attr("content")
	})
	err = c.Visit(constant.LoginUrl)
	return
}
