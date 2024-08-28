package engine

import (
	"github.com/zJiajun/warmane/database"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/model/table"
	"github.com/zJiajun/warmane/scraper"
	"gorm.io/gorm"
	"sync"
)

type Engine struct {
	scrapers *scraper.Scrapers
	db       *gorm.DB
	wg       sync.WaitGroup
}

func New() *Engine {
	db, err := database.Open()
	if err != nil {
		logger.Fatal(err)
	}
	if err = autoMigrate(db); err != nil {
		logger.Fatal(err)
	}
	e := &Engine{
		scrapers: scraper.New(db),
		db:       db,
	}
	e.init()
	return e
}

func (e *Engine) init() {
	accounts, err := e.ListOnlineAccount()
	if err != nil {
		logger.Error("init scrapers err:", err)
		return
	}
	for _, account := range accounts {
		e.scrapers.GetOrPut(account.AccountName)
		//go e.keepingAccountOnline(int64(account.ID))
	}
}

func (e *Engine) getScraper(name string) *scraper.Scraper {
	return e.scrapers.GetOrPut(name)
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&table.Account{},
		&table.AccountDetails{},
		&table.TradeInfo{},
	)
}
