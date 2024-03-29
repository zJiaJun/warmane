package storage

import (
	"gitub.com/zJiajun/warmane/model/table"
	"gorm.io/gorm"
	"net/url"
)

type SqliteStorage struct {
	name string
	db   *gorm.DB
}

func NewSqliteStorage(name string, db *gorm.DB) *SqliteStorage {
	return &SqliteStorage{
		name: name,
		db:   db,
	}
}

func (s *SqliteStorage) Init() error {
	return s.db.AutoMigrate(
		&table.Visited{},
		&table.Cookies{},
	)
}

func (s *SqliteStorage) Visited(requestID uint64) error {
	visited := &table.Visited{RequestID: int(requestID), Visited: 1}
	return s.db.Create(visited).Error
}

func (s *SqliteStorage) IsVisited(requestID uint64) (bool, error) {
	var count int64
	s.db.Model(&table.Visited{}).Where("request_id = ?", requestID).Count(&count)
	if count >= 1 {
		return true, nil
	}
	return false, nil
}

func (s *SqliteStorage) Cookies(u *url.URL) string {
	var cookies string
	s.db.Select("cookies").Model(&table.Cookies{}).Where("host = ?", u.Host).Where("name = ?", s.name).First(&cookies)
	return cookies
}

func (s *SqliteStorage) SetCookies(u *url.URL, cookies string) {
	ck := &table.Cookies{Host: u.Host, Name: s.name, Cookies: cookies}
	s.db.Create(ck)
}
