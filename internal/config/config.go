package config

import (
	"gitub.com/zJiajun/warmane/internal/errors"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	BaseUrl    = "https://www.warmane.com"
	AccountUrl = BaseUrl + "/account"
	LoginUrl   = AccountUrl + "/login"
	LogoutUrl  = AccountUrl + "/logout"
)
const (
	CsrfTokenSelector = "meta[name='csrf-token']"
	CoinsSelector     = ".myCoins"
	PointsSelector    = ".myPoints"
)

type (
	Config struct {
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

func LoadConf() (*Config, error) {
	file, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, errors.ErrConfNotFound
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, errors.ErrConfDecodeError
	}
	return &conf, nil
}
