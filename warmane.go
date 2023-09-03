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
	"io"
	"time"
)

const (
	baseUrl    = "https://www.warmane.com"
	accountUrl = baseUrl + "/account"
	loginUrl   = accountUrl + "/login"
	//see https://cn.2captcha.com/2captcha-api#solving_recaptchav2_new
	siteKey = "6LfXRRsUAAAAAEApnVwrtQ7aFprn4naEcc05AZUR"
)

func init() {
	flag.Parse()
}
func main() {
	defer glog.Flush()
	glog.Info("warmane collect daily point running")
	c := colly.NewCollector()
	c.SetRequestTimeout(5 * time.Second)
	//Add Random User agent
	extensions.RandomUserAgent(c)

	csrfToken := ""
	c.OnHTML("meta[name='csrf-token']", func(element *colly.HTMLElement) {
		csrfToken = element.Attr("content")
		glog.Info("warmane site csrf-token: ", csrfToken)
	})
	c.Visit(loginUrl)

	loginData := map[string]string{
		"return": "",
		"userID": "1",
		"userPW": "2",
	}
	code := solveReCaptcha()
	loginData["g-recaptcha-response"] = code

	c.OnRequest(func(request *colly.Request) {
		glog.Infof("warmane onRequest url %s", request.URL)
		request.Headers.Set("Origin", baseUrl)
		request.Headers.Set("Referer", loginUrl)
		request.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		request.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		request.Headers.Set("X-Csrf-Token", csrfToken)
		request.Headers.Set("X-Requested-With", "XMLHttpRequest")
		request.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		request.Headers.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	})
	c.OnResponse(func(response *colly.Response) {
		respBytes, err := decodeBody(response)
		if err != nil {
			glog.Error("decode login response body error: ", err)
		}
		glog.Infof("warmane login response statusCode %d, body %s", response.StatusCode, string(respBytes))
	})
	loginErr := c.Post(loginUrl, loginData)
	if loginErr != nil {
		glog.Error("warmane login error: ", loginErr)
	}

	//after login, visit warmane account url
	c.OnHTML(".wm-ui-hyper-custom-b", func(element *colly.HTMLElement) {
		attr := element.Attr("data-click")
		glog.Info("warmane collect points function: ", attr)
	})
	afterLoginErr := c.Visit(accountUrl)
	if afterLoginErr != nil {
		glog.Error("after login visit error: ", loginErr)
	}

}

func solveReCaptcha() string {
	client := api2captcha.NewClient("")
	client.DefaultTimeout = 120
	client.RecaptchaTimeout = 600
	client.PollingInterval = 30
	balance, err := client.GetBalance()
	if err != nil {
		glog.Error("captcha get balance error: ", err)
	}
	glog.Info("captcha account balance: ", balance)
	captcha := api2captcha.ReCaptcha{
		SiteKey: siteKey,
		Url:     loginUrl,
		Action:  "verify",
	}
	code, err := client.Solve(captcha.ToRequest())
	if err != nil {
		glog.Error("captcha solve error: ", err)
	}
	glog.Info("captcha return code: ", code)
	return code
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
