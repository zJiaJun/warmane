package storage

import (
	"github.com/zJiajun/warmane/constant"
	"github.com/zJiajun/warmane/model/table"
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
	return nil
}

func (s *SqliteStorage) Visited(requestID uint64) error {
	return nil
}

func (s *SqliteStorage) IsVisited(requestID uint64) (bool, error) {
	return false, nil
}

func (s *SqliteStorage) Cookies(u *url.URL) string {
	var cookies string
	s.db.Select("cookies").Model(&table.Account{}).Where("host = ?", u.Host).Where("account_name = ?", s.name).First(&cookies)
	return cookies
}

func (s *SqliteStorage) SetCookies(u *url.URL, cookies string) {
	/*
		ck := &table.Account{Host: u.Host, AccountName: s.name, Cookies: cookies}
		s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "host"}, {Name: "account_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"cookies"}),
		}).Create(ck)
	*/
}

func Clear(db *gorm.DB, name string) error {
	return db.Unscoped().Where("host = ?", constant.HOST).Where("account_name = ?", name).Delete(&table.Account{}).Error
}
