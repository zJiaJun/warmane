package engine

import (
	"github.com/zJiajun/warmane/config"
	"github.com/zJiajun/warmane/database"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/model/table"
	"github.com/zJiajun/warmane/scraper"
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
		scrapers: scraper.New(conf.Accounts, db),
		db:       db,
	}
}

func (e *Engine) getScraper(name string) *scraper.Scraper {
	return e.scrapers.Get(name)
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&table.DailyPoint{},
		&table.TradeInfo{},
	)
}
