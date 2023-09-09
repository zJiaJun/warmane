package main

import (
	"fmt"
	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/golang/glog"
	"time"
)

type captcha struct {
}

func (c *captcha) HandleCaptcha() (string, error) {
	client := api2captcha.NewClient(conf.CaptchaApiKey)
	client.DefaultTimeout = 120
	client.RecaptchaTimeout = 600
	client.PollingInterval = 30
	_, err := queryBalance(client)
	if err != nil {
		return "", fmt.Errorf("验证码破解服务查询余额失败, %w", err)
	}
	code, err := solveCaptcha(client)
	if err != nil {
		return "", fmt.Errorf("验证码破解服务执行失败, %w", err)
	}
	return code, nil
}

func queryBalance(client *api2captcha.Client) (float64, error) {
	balance, err := client.GetBalance()
	if err != nil {
		return 0.0, err
	}
	glog.Infof("验证码破解服务可用余额: %f美元", balance)
	return balance, nil
}

func solveCaptcha(client *api2captcha.Client) (string, error) {
	start := time.Now()
	glog.Info("验证码破解服务开始执行, 需等待1-2分钟")
	c := api2captcha.ReCaptcha{
		SiteKey: conf.WarmaneSiteKey,
		Url:     conf.Url.Login,
		Action:  "verify",
	}
	code, err := client.Solve(c.ToRequest())
	if err != nil {
		return "", err
	}
	glog.Infof("验证码破解服务执行成功, 耗时 %v", time.Since(start))
	return code, nil
}
