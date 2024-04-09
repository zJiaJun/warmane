package scraper

import (
	"github.com/zJiajun/warmane/config"
	"gorm.io/gorm"
)

type Scrapers struct {
	sc map[string]*Scraper
}

func New(accounts []config.Account, db *gorm.DB) *Scrapers {
	m := make(map[string]*Scraper, len(accounts))
	for _, v := range accounts {
		m[v.Username] = newScraper(v.Username, db)
	}
	return &Scrapers{sc: m}
}

func (s *Scrapers) Get(name string) *Scraper {
	return s.sc[name]
}
