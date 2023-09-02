package main

import (
	"flag"
	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/golang/glog"
	"time"
)

const (
	warmaneUrl = "https://www.warmane.com"
	loginUrl   = warmaneUrl + "/account/login"
)

func init() {
	flag.Parse()
}
func main() {
	glog.Info("warmane collect daily point running")
	loginData := map[string]string{
		"return":               "",
		"userID":               "1",
		"userPW":               "2",
		"g-recaptcha-response": "",
	}
	defer glog.Flush()

	c := colly.NewCollector()
	c.SetRequestTimeout(5 * time.Second)
	requestInit(c)
	c.OnResponse(func(response *colly.Response) {
		glog.Infof("warmane onResponse statusCode %d, body %s",
			response.StatusCode, string(response.Body))
	})
	c.OnError(func(response *colly.Response, err error) {
		glog.Errorf("warmane login error %s", err.Error())
	})
	err := c.Post(loginUrl, loginData)
	if err != nil {
		glog.Error("login post err %v", err)
	}
	//c.Wait()

}

func requestInit(c *colly.Collector) {
	//Add Random User agent
	extensions.RandomUserAgent(c)

	c.OnRequest(func(request *colly.Request) {
		glog.Infof("warmane onRequest url %s", request.URL)
		request.Headers.Set("Origin", warmaneUrl)
		request.Headers.Set("Referer", loginUrl)
	})
}

func captchaPass() {
	client := api2captcha.NewClient("YOUR_API_KEY")
	client.ApiKey = ""
}
