package scraper

import "gorm.io/gorm"

type Scrapers struct {
	sc map[string]*Scraper
	db *gorm.DB
}

func New(db *gorm.DB) *Scrapers {
	return &Scrapers{
		sc: make(map[string]*Scraper),
		db: db,
	}
}

func (s *Scrapers) GetOrPut(name string) *Scraper {
	if _, ok := s.sc[name]; !ok {
		s.sc[name] = newScraper(name, s.db)
	}
	return s.sc[name]
}
