package storage

import (
	"github.com/gocolly/colly/v2/storage"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/errors"
	"gitub.com/zJiajun/warmane/logger"
	"net/url"
	"os"
)

type DiskStorage struct {
	fileName        string
	inMemoryStorage *storage.InMemoryStorage
}

func NewDiskStorage(name string) *DiskStorage {
	return &DiskStorage{
		fileName:        constant.CookieFileName(name),
		inMemoryStorage: &storage.InMemoryStorage{},
	}
}

func (ds *DiskStorage) Init() error {
	return ds.inMemoryStorage.Init()
}

func (ds *DiskStorage) Visited(requestID uint64) error {
	return ds.inMemoryStorage.Visited(requestID)
}

func (ds *DiskStorage) IsVisited(requestID uint64) (bool, error) {
	return ds.inMemoryStorage.IsVisited(requestID)
}

func (ds *DiskStorage) Cookies(u *url.URL) string {
	buf, err := os.ReadFile(ds.fileName)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (ds *DiskStorage) SetCookies(u *url.URL, cookies string) {
	if err := os.WriteFile(ds.fileName, []byte(cookies), 0644); err != nil {
		logger.Error(err)
	}
}

func deleteCookies(cookiesFile string) {
	_, err := os.Stat(cookiesFile)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(cookiesFile)
	logger.Infof("删除历史[%s]文件", cookiesFile)
}

func validateCookies(cookiesFile string) error {
	_, err := os.Stat(cookiesFile)
	if os.IsNotExist(err) {
		logger.Errorf("不存在[%s]文件", cookiesFile)
		return errors.ErrCookieNotFound
	}
	file, err := os.ReadFile(cookiesFile)
	if err != nil {
		return err
	}
	cookies := storage.UnstringifyCookies(string(file))
	for _, v := range constant.CookieKeys {
		if !storage.ContainsCookie(cookies, v) {
			logger.Infof("存在[%s]文件,不匹配cookieKey[%s]", cookiesFile, v)
			//_ = os.Remove(cookiesFile)
			return errors.ErrCookieNonMatchKey
		}
	}
	return nil
}
