package scraper

import "gitub.com/zJiajun/warmane/config"

type Scrapers struct {
	sc map[string]*Scraper
}

func New(accounts []config.Account) *Scrapers {
	m := make(map[string]*Scraper, len(accounts))
	for _, v := range accounts {
		m[v.Username] = newScraper(v.Username)
	}
	return &Scrapers{sc: m}
}

func (s *Scrapers) Get(name string) *Scraper {
	return s.sc[name]
}
