package storage

import (
	"github.com/gocolly/colly/v2/storage"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/errors"
	"gitub.com/zJiajun/warmane/logger"
	"gitub.com/zJiajun/warmane/model/table"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "host"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"cookies"}),
	}).Create(ck)
}

func Validate(db *gorm.DB, name string) error {
	var cookiesInDB string
	db.Select("cookies").Model(&table.Cookies{}).Where("host = ?", constant.HOST).Where("name = ?", name).First(&cookiesInDB)
	if cookiesInDB == "" {
		return errors.ErrCookieNotFound
	} else {
		cookies := storage.UnstringifyCookies(cookiesInDB)
		for _, v := range constant.CookieKeys {
			if !storage.ContainsCookie(cookies, v) {
				logger.Infof("cookies数据存在,但不匹配cookieKey[%s]", v)
				return errors.ErrCookieNonMatchKey
			}
		}
	}
	return nil
}
