package engine

import (
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/scraper"
	"sync"
)

type Engine struct {
	config   *config.Config
	scrapers *scraper.Scrapers
	wg       sync.WaitGroup
}

func New(cfg string) *Engine {
	conf, err := config.Load(cfg)
	if err != nil {
		panic(err)
	}
	return &Engine{
		config:   conf,
		scrapers: scraper.New(conf.Accounts),
	}
}

func (e *Engine) getScraper(name string) *scraper.Scraper {
	return e.scrapers.Get(name)
}
