package captcha

import (
	"fmt"
	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/internal/config"
	"time"
)

type Captcha struct {
	siteKey string
	url     string
	client  *api2captcha.Client
}

func NewCaptcha(captchaApiKey string, siteKey string) *Captcha {
	client := api2captcha.NewClient(captchaApiKey)
	client.DefaultTimeout = 120
	client.RecaptchaTimeout = 600
	client.PollingInterval = 30
	return &Captcha{
		siteKey: siteKey,
		url:     config.LoginUrl,
		client:  client,
	}
}

func (c *Captcha) HandleCaptcha() (string, error) {
	if _, err := c.queryBalance(); err != nil {
		return "", fmt.Errorf("验证码破解服务查询余额失败, %w", err)
	}
	code, err := c.solveCaptcha()
	if err != nil {
		return "", fmt.Errorf("验证码破解服务执行失败, %w", err)
	}
	return code, nil
}

func (c *Captcha) queryBalance() (float64, error) {
	balance, err := c.client.GetBalance()
	if err != nil {
		return 0.0, err
	}
	glog.Infof("验证码破解服务可用余额: %f美元", balance)
	return balance, nil
}

func (c *Captcha) solveCaptcha() (string, error) {
	start := time.Now()
	glog.Info("验证码破解服务开始执行, 需等待1-2分钟")
	r := api2captcha.ReCaptcha{
		SiteKey: c.siteKey,
		Url:     c.url,
		Action:  "verify",
	}
	code, err := c.client.Solve(r.ToRequest())
	if err != nil {
		return "", err
	}
	glog.Infof("验证码破解服务执行成功, 耗时 %v", time.Since(start))
	return code, nil
}
