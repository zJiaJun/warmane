package engine

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/zJiajun/warmane/captcha"
	"github.com/zJiajun/warmane/config"
	"github.com/zJiajun/warmane/constant"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/scraper/storage"
	"strings"
	"time"
)

func (e *Engine) KeepSession() {
	t := time.NewTicker(15 * time.Minute)
	for ; true; <-t.C {
		if err := e.login(e.config.Accounts[0]); err != nil {
			logger.Errorf("账号[%s]登录错误, 原因: %v", e.config.Accounts[0].Username, err)
		}
	}
	/*
		for {
			select {
			case <-t.C:
				if err := e.login(e.config.Accounts[0]); err != nil {
					logger.Errorf("账号[%s]登录错误, 原因: %v", e.config.Accounts[0].Username, err)
				}
			}
		}
	*/
}

func (e *Engine) login(account *config.Account) error {
	name := account.Username
	if e.config.UseCookiesLogin {
		logger.Infof("配置项[useCookiesLogin]为true, 使用cookies登录")
		if err := storage.Validate(e.db, name); err == nil {
			logger.Infof("[%s]存在cookies且验证通过,使用cookies登录", name)
		} else {
			return err
		}
	} else {
		logger.Infof("配置项[useCookiesLogin]为false, 使用2captcha方式登录")
		capt := captcha.NewCaptcha(e.config.CaptchaApiKey, e.config.WarmaneSiteKey, constant.LoginUrl)
		loginData := make(map[string]string, 5)
		if code, err := capt.HandleCaptcha(); err == nil {
			loginData["return"] = ""
			loginData["userID"] = name
			loginData["userPW"] = account.Password
			loginData["g-recaptcha-response"] = code
			loginData["userRM"] = "on"
		} else {
			return err
		}
		_ = storage.Clear(e.db, name)
		c := e.getScraper(name).CloneCollector()
		e.getScraper(name).SetRequestHeaders(c)
		e.getScraper(name).DecodeResponse(c)
		var bodyErr error
		var bodyMsg struct {
			Messages struct {
				Success []string `json:"success"`
				Error   []string `json:"error"`
			}
		}
		c.OnResponse(func(response *colly.Response) {
			bodyText := string(response.Body)
			bodyErr = json.Unmarshal(response.Body, &bodyMsg)
			if bodyErr != nil {
				bodyErr = fmt.Errorf("账号[%s]登陆解码Json错误, 返回内容: %s", name, bodyText)
				return
			}
			if len(bodyMsg.Messages.Error) > 0 {
				errMsg := bodyMsg.Messages.Error[0]
				bodyErr = fmt.Errorf("账号[%s]登录失败, %s", name, errMsg)
			}
		})
		err := c.Post(constant.LoginUrl, loginData)
		if bodyErr != nil {
			return bodyErr
		}
		if err != nil {
			return err
		}
	}
	isLogin, isAuth := e.isLogin(account)
	if isLogin {
		logger.Infof("账号[%s]登录成功", name)
	} else if isAuth {
		logger.Infof("账号[%s]二次认证(Two factor authentication)开始", name)
		e.auth(account)
	} else {
		return fmt.Errorf("账号[%s]未登录", name)
	}
	e.config.UseCookiesLogin = true
	return nil
}

func (e *Engine) logout(account *config.Account) error {
	c := e.getScraper(account.Username).CloneCollector()
	err := c.Visit(constant.LogoutUrl)
	if err != nil {
		return err
	}
	logger.Infof("账号[%s]退出成功", account.Username)
	return err
}

func (e *Engine) isLogin(account *config.Account) (bool, bool) {
	isLogin, isAuth := false, false
	c := e.getScraper(account.Username).CloneCollector()
	e.getScraper(account.Username).SetRequestHeaders(c)
	e.getScraper(account.Username).DecodeResponse(c)
	c.OnHTML("div.content-inner.left > table > tbody > tr:nth-child(2) > td", func(element *colly.HTMLElement) {
		isLogin = strings.Contains(element.Text, account.Username)
	})
	c.OnHTML("form[id='frmAuthenticate']", func(element *colly.HTMLElement) {
		isAuth = true
	})
	_ = c.Visit(constant.AccountUrl)
	return isLogin, isAuth
}

func (e *Engine) auth(account *config.Account) {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	var authCode string
	logger.Infof("输入二次认证码[Auth Code]回车键结束")
	_, _ = fmt.Scanln(&authCode)
	authData := map[string]string{"authCode": authCode}
	_ = c.Post(constant.AuthenticationUrl, authData)
}
