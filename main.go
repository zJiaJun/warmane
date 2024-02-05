package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"github.com/andybalholm/brotli"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/golang/glog"
	"github.com/klauspost/compress/flate"
	"gitub.com/zJiajun/warmane/captcha"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"sync"
	"time"
)

type (
	Config struct {
		Url struct {
			Base    string `yaml:"base"`
			Account string `yaml:"account"`
			Login   string `yaml:"login"`
			Logout  string `yaml:"logout"`
		}
		Selector struct {
			CsrfToken string `yaml:"csrfToken"`
			Coins     string `yaml:"coins"`
			Points    string `yaml:"points"`
		}
		CaptchaApiKey  string    `yaml:"captchaApiKey"`
		WarmaneSiteKey string    `yaml:"warmaneSiteKey"`
		Accounts       []Account `yaml:"accounts"`
	}
	Account struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
)

type BodyMsg struct {
	Messages struct {
		Success []string `json:"success"`
		Error   []string `json:"error"`
	}
	Points []float64 `json:"points"`
}

var (
	conf            Config
	bodyMsg         BodyMsg
	wg              sync.WaitGroup
	csrfTokenError  = errors.New("查询获取csrfToken错误")
	confNotFound    = errors.New("配置文件[config.yml]未找到, 请把配置文件放到程序同一目录下")
	confDecodeError = errors.New("配置文件[config.yml]解析错误, 请检查配置文件")
)

const (
	/*
			{"messages":{"error":["Incorrect account name or password."]}}
			{"messages":{"error":["You have already collected your points today."]}}
		 	{"messages":{"error":["You have not logged in-game today."]}}
			{"messages":{"success":["Daily points collected."]},"points":[10.4]}
	*/
	loginSuccessBody = "{\"redirect\":[\"\\/account\"]}"
)

func init() {
	flag.Parse()
}

func loadConf() error {
	file, err := os.ReadFile("conf.yml")
	if err != nil {
		return confNotFound
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return confDecodeError
	}
	return nil
}

func main() {
	glog.Info("开始运行自动签到功能")
	defer glog.Flush()
	err := loadConf()
	if err != nil {
		handleError(err)
		return
	}
	count := len(conf.Accounts)
	glog.Infof("加载配置文件[conf.yml]成功, 需要签到的账号数量是[%d]", count)
	wg.Add(count)
	glog.Infof("开始goroutine并发处理")
	for _, account := range conf.Accounts {
		go loginAndCollect(account)
	}
	wg.Wait()
}

func handleError(err error) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, confNotFound):
		glog.Error(err.Error())
	case errors.Is(err, confDecodeError):
		glog.Error(err.Error())
	case errors.Is(err, csrfTokenError):
		glog.Error(err.Error())
	default:
		glog.Error(err.Error())
	}
}

func loginAndCollect(account Account) {
	defer wg.Done()
	c := colly.NewCollector()
	//允许重复访问URL
	c.AllowURLRevisit = true
	c.SetRequestTimeout(5 * time.Second)
	//Add Random User agent
	extensions.RandomUserAgent(c)

	csrfToken, err := queryCsrfToken(c)
	if err != nil {
		handleError(err)
		return
	}

	loginSuccess := false
	loginData := make(map[string]string, 4)
	loginData["return"] = ""
	loginData["userID"] = account.Username
	loginData["userPW"] = account.Password

	capt := captcha.Captcha{
		CaptchaApiKey:  conf.CaptchaApiKey,
		WarmaneSiteKey: conf.WarmaneSiteKey,
		LoginUrl:       conf.Url.Login,
	}
	code, err := capt.HandleCaptcha()
	if err != nil {
		handleError(err)
		return
	}
	loginData["g-recaptcha-response"] = code

	requestCallback := func(request *colly.Request) {
		request.Headers.Set("Origin", conf.Url.Base)
		request.Headers.Set("Referer", conf.Url.Login)
		request.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		request.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		request.Headers.Set("X-Csrf-Token", csrfToken)
		request.Headers.Set("X-Requested-With", "XMLHttpRequest")
		request.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		request.Headers.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	}
	c.OnRequest(requestCallback)

	responseCallback := func(response *colly.Response) {
		respBytes, decodeErr := decodeBody(response)
		if decodeErr != nil {
			glog.Errorf("解码[%s]返回内容错误, %v", response.Request.URL, err)
			return
		}
		//需要将解码后的body赋值回去, 否则下面的onHTML无法解析selector
		response.Body = respBytes
		glog.Infof("解码[%s]返回内容成功, 状态码:[%d], 大小[%d]", response.Request.URL, response.StatusCode, len(respBytes))
		if conf.Url.Login == response.Request.URL.String() && response.Request.Method == "POST" {
			bodyText := string(response.Body)
			if bodyText == loginSuccessBody {
				loginSuccess = true
				glog.Infof("账号[%s]登录成功", account.Username)
			} else {
				err := json.Unmarshal(response.Body, &bodyMsg)
				if err != nil {
					glog.Errorf("账号[%s]登陆解码Json错误, 返回内容: %s", account.Username, bodyText)
					return
				}
				if len(bodyMsg.Messages.Error) > 0 {
					errMsg := bodyMsg.Messages.Error[0]
					glog.Errorf("账号[%s]登录失败, %s", account.Username, errMsg)
				} else {
					glog.Errorf("账号[%s]登录失败, %s", account.Username, bodyText)
				}
			}
		}
		if conf.Url.Account == response.Request.URL.String() && response.Request.Method == "POST" {
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
		}
	}
	c.OnResponse(responseCallback)

	loginErr := c.Post(conf.Url.Login, loginData)
	if loginErr != nil {
		glog.Errorf("账号[%s]登录错误: %v", account.Username, loginErr)
		return
	}
	if loginSuccess {
		coins := ""
		points := ""
		c.OnHTML(conf.Selector.Coins, func(element *colly.HTMLElement) {
			coins = element.Text
		})
		c.OnHTML(conf.Selector.Points, func(element *colly.HTMLElement) {
			points = element.Text
		})
		accErr := c.Visit(conf.Url.Account)
		if accErr != nil {
			glog.Errorf("账号[%s]访问账号页面错误: %v", account.Username, accErr)
			return
		}
		glog.Infof("账号[%s]收集签到点[前]的 coins: [%s]", account.Username, coins)
		glog.Infof("账号[%s]收集签到点[前]的 points: [%s]", account.Username, points)
		collectPointsData := map[string]string{
			"collectpoints": "true",
		}
		accErr = c.Post(conf.Url.Account, collectPointsData)
		if accErr != nil {
			glog.Errorf("账号[%s]收集签到点错误: %v", account.Username, accErr)
			return
		}
		accErr = c.Visit(conf.Url.Account)
		if accErr != nil {
			glog.Errorf("账号[%s]访问账号页面错误: %v", account.Username, accErr)
			return
		}
		glog.Infof("账号[%s]收集签到点[后]的 coins: [%s]", account.Username, coins)
		glog.Infof("账号[%s]收集签到点[后]的 points: [%s]", account.Username, points)
		accErr = c.Visit(conf.Url.Logout)
		if accErr != nil {
			glog.Errorf("账号[%s]退出错误: %v", account.Username, accErr)
			return
		}
		glog.Infof("账号[%s]退出成功", account.Username)
	}
}

func queryCsrfToken(c *colly.Collector) (string, error) {
	csrfToken := ""
	csrfTokenCallback := func(element *colly.HTMLElement) {
		csrfToken = element.Attr("content")
	}
	c.OnHTML(conf.Selector.CsrfToken, csrfTokenCallback)
	err := c.Visit(conf.Url.Login)
	if err != nil {
		return "", csrfTokenError
	}
	c.OnHTMLDetach(conf.Selector.CsrfToken)
	if csrfToken == "" {
		return "", csrfTokenError
	}
	glog.Infof("查询获取warmane网站的csrfToken成功: %s", csrfToken)
	return csrfToken, nil
}

func decodeBody(response *colly.Response) ([]byte, error) {
	encoding := response.Headers.Get("Content-Encoding")
	responseBody := response.Body
	switch encoding {
	case "br":
		return io.ReadAll(brotli.NewReader(bytes.NewBuffer(responseBody)))
	case "gzip":
		gr, _ := gzip.NewReader(bytes.NewBuffer(responseBody))
		return io.ReadAll(gr)
	case "deflate":
		zr := flate.NewReader(bytes.NewBuffer(responseBody))
		defer func(zr io.ReadCloser) {
			_ = zr.Close()
		}(zr)
		return io.ReadAll(zr)
	default:
		return io.ReadAll(bytes.NewBuffer(responseBody))
	}
}
