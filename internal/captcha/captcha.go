package captcha

import (
	"fmt"
	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/internal/config"
	"time"
)

type Captcha struct {
	captchaApiKey  string
	warmaneSiteKey string
	loginUrl       string
}

func New(conf *config.Config) *Captcha {
	return &Captcha{
		captchaApiKey:  conf.CaptchaApiKey,
		warmaneSiteKey: conf.WarmaneSiteKey,
		loginUrl:       config.LoginUrl,
	}
}

func (c *Captcha) HandleCaptcha() (string, error) {
	client := api2captcha.NewClient(c.captchaApiKey)
	client.DefaultTimeout = 120
	client.RecaptchaTimeout = 600
	client.PollingInterval = 30
	if _, err := queryBalance(client); err != nil {
		return "", fmt.Errorf("验证码破解服务查询余额失败, %w", err)
	}
	/*
		code, err := solveCaptcha(client, c.warmaneSiteKey, c.loginUrl)
		if err != nil {
			return "", fmt.Errorf("验证码破解服务执行失败, %w", err)
		}
	*/
	return "code", nil
}

func queryBalance(client *api2captcha.Client) (float64, error) {
	balance, err := client.GetBalance()
	if err != nil {
		return 0.0, err
	}
	glog.Infof("验证码破解服务可用余额: %f美元", balance)
	return balance, nil
}

func solveCaptcha(client *api2captcha.Client, warmaneSiteKey string, loginUrl string) (string, error) {
	start := time.Now()
	glog.Info("验证码破解服务开始执行, 需等待1-2分钟")
	c := api2captcha.ReCaptcha{
		SiteKey: warmaneSiteKey,
		Url:     loginUrl,
		Action:  "verify",
	}
	code, err := client.Solve(c.ToRequest())
	if err != nil {
		return "", err
	}
	glog.Infof("验证码破解服务执行成功, 耗时 %v", time.Since(start))
	return code, nil
}
