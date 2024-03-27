package engine

import (
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/database"
	"gitub.com/zJiajun/warmane/logger"
	"gitub.com/zJiajun/warmane/model"
	"gitub.com/zJiajun/warmane/scraper"
	"gorm.io/gorm"
	"sync"
)

type Engine struct {
	config   *config.Config
	scrapers *scraper.Scrapers
	db       *gorm.DB
	wg       sync.WaitGroup
}

func New(cfg string) *Engine {
	conf, err := config.Load(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	db, err := database.Open()
	if err != nil {
		logger.Fatal(err)
	}
	if err = autoMigrate(db); err != nil {
		logger.Fatal(err)
	}
	return &Engine{
		config:   conf,
		scrapers: scraper.New(conf.Accounts),
		db:       db,
	}
}

func (e *Engine) getScraper(name string) *scraper.Scraper {
	return e.scrapers.Get(name)
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.TradeInfo{},
	)
}
