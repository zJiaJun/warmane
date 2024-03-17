package engine

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/storage"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/captcha"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/errors"
	"gitub.com/zJiajun/warmane/model"
	"os"
	"strings"
)

func (e *Engine) login(account config.Account) error {
	name := account.Username
	cookiesFile := constant.CookieFileName(name)
	if e.config.UseCookiesLogin {
		glog.Infof("配置项[useCookiesLogin]为true, 使用[%s]文件登录", cookiesFile)
		if err := validateCookies(cookiesFile); err == nil {
			glog.Infof("存在[%s]文件且验证通过,使用cookies文件登录", cookiesFile)
		} else {
			return err
		}
	} else {
		glog.Infof("配置项[useCookiesLogin]为false, 使用2captcha方式登录")
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
		deleteCookies(cookiesFile)
		c := e.getScraper(name).CloneCollector()
		e.getScraper(name).SetRequestHeaders(c)
		e.getScraper(name).DecodeResponse(c)
		var bodyErr error
		var bodyMsg model.BodyMsg
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
		glog.Infof("账号[%s]登录成功", name)
	} else if isAuth {
		glog.Infof("账号[%s]二次认证(Two factor authentication)开始", name)
		e.auth(account)
	} else {
		return fmt.Errorf("账号[%s]未登录", name)
	}
	return nil
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

func (e *Engine) isLogin(account config.Account) (bool, bool) {
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

func (e *Engine) auth(account config.Account) {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	var authCode string
	glog.Infof("输入二次认证码[Auth Code]回车键结束")
	fmt.Scanln(&authCode)
	authData := map[string]string{"authCode": authCode}
	_ = c.Post(constant.AuthenticationUrl, authData)
}

func deleteCookies(cookiesFile string) {
	_, err := os.Stat(cookiesFile)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(cookiesFile)
	glog.Infof("删除历史[%s]文件", cookiesFile)
}

func validateCookies(cookiesFile string) error {
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
