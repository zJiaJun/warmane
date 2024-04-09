package config

import (
	"github.com/zJiajun/warmane/errors"
	"gopkg.in/yaml.v3"
	"os"
)

type (
	Config struct {
		CaptchaApiKey   string    `yaml:"captchaApiKey"`
		WarmaneSiteKey  string    `yaml:"warmaneSiteKey"`
		UseCookiesLogin bool      `yaml:"useCookiesLogin"`
		Accounts        []Account `yaml:"accounts"`
	}
	Account struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
)

func Load(cfg string) (*Config, error) {
	file, err := os.ReadFile(cfg)
	if err != nil {
		return nil, errors.ErrConfNotFound
	}
	var conf Config
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, errors.ErrConfDecodeError
	}
	return &conf, nil
}
