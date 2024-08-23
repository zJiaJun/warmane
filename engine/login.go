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
)

func (e *Engine) login(account *config.Account) error {
	name := account.Username
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
