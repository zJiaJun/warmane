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
	siteKey              = "6LfXRRsUAAAAAEApnVwrtQ7aFprn4naEcc05AZUR"
	csrfTokenSelector    = "meta[name='csrf-token']"
	coinsSelector        = ".myCoins"
	pointsSelector       = ".myPoints"
	loginSuccessBodyText = "{\"redirect\":[\"\\/account\"]}"
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
	csrfTokenCallback := func(element *colly.HTMLElement) {
		csrfToken = element.Attr("content")
		glog.Info("warmane site csrf-token: ", csrfToken)
	}
	c.OnHTML(csrfTokenSelector, csrfTokenCallback)
	err := c.Visit(loginUrl)
	if err != nil {
		glog.Errorf("warmane visit %s error %v", loginUrl, err)
		return
	}
	c.OnHTMLDetach(csrfTokenSelector)

	loginData := map[string]string{
		"return": "",
		"userID": "1",
		"userPW": "2",
	}
	code := handleCaptcha()
	loginData["g-recaptcha-response"] = code

	requestCallback := func(request *colly.Request) {
		glog.Infof("[requestCallback] warmane %s onRequest", request.URL)
		request.Headers.Set("Origin", baseUrl)
		request.Headers.Set("Referer", loginUrl)
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
		if loginUrl == response.Request.URL.String() && response.Request.Method == "POST" {
			bodyText := string(response.Body)
			if bodyText == loginSuccessBodyText {
				glog.Info("[responseCallback] warmane login success")
			} else {
				glog.Error("[responseCallback] warmane login failure: ", bodyText)
			}
		}
		if accountUrl == response.Request.URL.String() && response.Request.Method == "POST" {
			bodyText := string(response.Body)
			glog.Info("[responseCallback] warmane collect points body: ", bodyText)
		}
	}
	c.OnResponse(responseCallback)

	loginErr := c.Post(loginUrl, loginData)
	if loginErr != nil {
		glog.Error("warmane login error: ", loginErr)
		return
	}

	c.OnHTML(coinsSelector, func(element *colly.HTMLElement) {
		coins := element.Text
		glog.Info("warmane account coins: ", coins)
	})
	c.OnHTML(pointsSelector, func(element *colly.HTMLElement) {
		points := element.Text
		glog.Info("warmane account points: ", points)
	})
	accountUrlErr := c.Visit(accountUrl)
	if accountUrlErr != nil {
		glog.Error("after login visit error: ", accountUrlErr)
	}
	collectPointsData := map[string]string{
		"collectpoints": "true",
	}
	collectPointsErr := c.Post(accountUrl, collectPointsData)
	if collectPointsErr != nil {
		glog.Error("collect points error: ", collectPointsErr)
	}
	c.Visit(accountUrl)
}

func handleCaptcha() string {
	client := api2captcha.NewClient("")
	client.DefaultTimeout = 120
	client.RecaptchaTimeout = 600
	client.PollingInterval = 30

	b1 := queryBalance(client)
	code := solveCaptcha(client)
	b2 := queryBalance(client)
	glog.Info("2captcha solve captcha cost: ", b2-b1)
	return code
}

func queryBalance(client *api2captcha.Client) float64 {
	balance, err := client.GetBalance()
	if err != nil {
		glog.Error("2captcha get balance error: ", err)
	}
	glog.Info("2captcha account balance: ", balance)
	return balance
}

func solveCaptcha(client *api2captcha.Client) string {
	glog.Info("2captcha solve captcha begin, waiting......")
	defer logElapsedTime("2captcha solve captcha finish", time.Now())
	captcha := api2captcha.ReCaptcha{
		SiteKey: siteKey,
		Url:     loginUrl,
		Action:  "verify",
	}
	code, err := client.Solve(captcha.ToRequest())
	if err != nil {
		glog.Error("2captcha solve error: ", err)
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
