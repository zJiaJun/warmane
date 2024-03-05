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
	"os"
)

func (e *Engine) login(account config.Account) error {
	name := account.Username
	cookiesFile := constant.CookieFileName(name)
	if e.config.UseCookiesLogin {
		glog.Infof("配置项[useCookiesLogin]为true, 使用[%s]文件登录", cookiesFile)
		if err := validateCookies(cookiesFile); err != nil {
			return err
		} else {
			glog.Infof("存在[%s]文件且验证通过,使用cookies文件登录", cookiesFile)
			return nil
		}
	} else {
		glog.Infof("配置项[useCookiesLogin]为false, 使用2captcha方式登录")
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
		if err = deleteCookies(cookiesFile); err != nil {
			return err
		}

		c := e.getScraper(name).CloneCollector()
		e.getScraper(name).SetRequestHeaders(c)
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

func (e *Engine) logout(account config.Account) error {
	c := e.getScraper(account.Username).CloneCollector()
	err := c.Visit(constant.LogoutUrl)
	if err != nil {
		return err
	}
	glog.Infof("账号[%s]退出成功", account.Username)
	return err
}

func deleteCookies(cookiesFile string) error {
	_, err := os.Stat(cookiesFile)
	if os.IsNotExist(err) {
		return nil
	}
	err = os.Remove(cookiesFile)
	glog.Infof("删除历史[%s]文件", cookiesFile)
	return err
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
