package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/andybalholm/brotli"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/golang/glog"
	"github.com/klauspost/compress/flate"
	"gopkg.in/yaml.v3"
	"io"
	"os"
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

var conf Config

const (
	loginSuccessBody         = "{\"redirect\":[\"\\/account\"]}"
	incorrectLoginBody       = "{\"messages\":{\"error\":[\"Incorrect account name or password.\"]}}"
	alreadyCollectPointsBody = "{\"messages\":{\"error\":[\"You have already collected your points today.\"]}}"
)

func init() {
	flag.Parse()
}

func loadConf() {
	file, err := os.ReadFile("conf.yml")
	if err != nil {
		glog.Error("can't find conf.yml file")
		return
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		glog.Error("load conf.yml error, ", err)
		return
	}
	glog.Info("load conf.yml success, ", conf)
}

func main() {
	loadConf()
	defer glog.Flush()
	glog.Info("warmane collect daily point running")
	c := colly.NewCollector()
	//允许重复访问URL
	c.AllowURLRevisit = true
	c.SetRequestTimeout(5 * time.Second)
	//Add Random User agent
	extensions.RandomUserAgent(c)

	csrfToken := ""
	csrfTokenCallback := func(element *colly.HTMLElement) {
		csrfToken = element.Attr("content")
		glog.Info("warmane site csrf-token: ", csrfToken)
	}
	c.OnHTML(conf.Selector.CsrfToken, csrfTokenCallback)
	err := c.Visit(conf.Url.Login)
	if err != nil {
		glog.Errorf("warmane visit %s error %v", conf.Url.Login, err)
		return
	}
	c.OnHTMLDetach(conf.Selector.CsrfToken)

	loginSuccuss := false
	loginData := make(map[string]string, 4)
	loginData["return"] = ""
	loginData["userID"] = conf.Accounts[0].Username
	loginData["userPW"] = conf.Accounts[0].Password
	code := handleCaptcha()
	if code == "" {
		glog.Error("2captcha error")
		return
	}
	loginData["g-recaptcha-response"] = code

	requestCallback := func(request *colly.Request) {
		glog.Infof("[requestCallback] warmane %s onRequest", request.URL)
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

	/*
		responseHeadersCallback := func(response *colly.Response) {
			if loginUrl == response.Request.URL.String() {
				respCookies := response.Headers.Values("Set-Cookie")
				glog.Info("[responseHeadersCallback] warmane get cookie: ", respCookies)
			}
		}
		c.OnResponseHeaders(responseHeadersCallback)
	*/

	responseCallback := func(response *colly.Response) {
		respBytes, err := decodeBody(response)
		if err != nil {
			glog.Errorf("[responseCallback] warmane %s decode response body error %v",
				response.Request.URL, err)
			return
		}
		//需要将解码后的body赋值回去, 否则下面的onHTML无法解析selector
		response.Body = respBytes
		glog.Infof("[responseCallback] warmane %s response statusCode %d, body size %d",
			response.Request.URL, response.StatusCode, len(respBytes))
		if conf.Url.Login == response.Request.URL.String() && response.Request.Method == "POST" {
			bodyText := string(response.Body)
			if bodyText == loginSuccessBody {
				loginSuccuss = true
				glog.Info("[responseCallback] warmane login success")
			} else {
				glog.Error("[responseCallback] warmane login failure: ", bodyText)
			}
		}
		if conf.Url.Account == response.Request.URL.String() && response.Request.Method == "POST" {
			bodyText := string(response.Body)
			glog.Info("[responseCallback] warmane collect points body: ", bodyText)
		}
	}
	c.OnResponse(responseCallback)

	loginErr := c.Post(conf.Url.Login, loginData)
	if loginErr != nil {
		glog.Error("warmane login error: ", loginErr)
		return
	}
	if loginSuccuss {
		c.OnHTML(conf.Selector.Coins, func(element *colly.HTMLElement) {
			coins := element.Text
			glog.Info("warmane account coins: ", coins)
		})
		c.OnHTML(conf.Selector.Points, func(element *colly.HTMLElement) {
			points := element.Text
			glog.Info("warmane account points: ", points)
		})
		accountUrlErr := c.Visit(conf.Url.Account)
		if accountUrlErr != nil {
			glog.Error("after login visit error: ", accountUrlErr)
			return
		}
		collectPointsData := map[string]string{
			"collectpoints": "true",
		}
		collectPointsErr := c.Post(conf.Url.Account, collectPointsData)
		if collectPointsErr != nil {
			glog.Error("collect points error: ", collectPointsErr)
			return
		}
		err := c.Visit(conf.Url.Account)
		if err != nil {
			return
		}
		err = c.Visit(conf.Url.Logout)
		if err != nil {
			return
		}
	}
}

func handleCaptcha() string {
	client := api2captcha.NewClient(conf.CaptchaApiKey)
	client.DefaultTimeout = 120
	client.RecaptchaTimeout = 600
	client.PollingInterval = 30

	code := ""
	b1 := queryBalance(client)
	if b1 > 0 {
		code = solveCaptcha(client)
		if len(code) > 0 {
			b2 := queryBalance(client)
			glog.Info("2captcha solve captcha cost: ", b2-b1)
		}
	}
	return code
}

func queryBalance(client *api2captcha.Client) float64 {
	balance, err := client.GetBalance()
	if err != nil {
		glog.Error("2captcha get balance error, ", err)
		return 0
	}
	glog.Info("2captcha account balance: ", balance)
	return balance
}

func solveCaptcha(client *api2captcha.Client) string {
	glog.Info("2captcha solve captcha begin, waiting......")
	defer logElapsedTime("2captcha solve captcha finish", time.Now())
	captcha := api2captcha.ReCaptcha{
		SiteKey: conf.WarmaneSiteKey,
		Url:     conf.Url.Login,
		Action:  "verify",
	}
	code, err := client.Solve(captcha.ToRequest())
	if err != nil {
		glog.Error("2captcha solve error, ", err)
		return ""
	}
	glog.Info("2captcha return code: ", code)
	return code
}

func logElapsedTime(msg string, start time.Time) {
	duration := time.Since(start)
	glog.Info(msg+" duration: ", duration)
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
		defer zr.Close()
		return io.ReadAll(zr)
	default:
		return io.ReadAll(bytes.NewBuffer(responseBody))
	}
}
