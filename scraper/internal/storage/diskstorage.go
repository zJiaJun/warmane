package storage

import (
	"github.com/gocolly/colly/v2/storage"
	"gitub.com/zJiajun/warmane/constant"
	"log"
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
	//glog.Infof("disk storage run cookies, %v", u)
	buf, err := os.ReadFile(ds.fileName)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (ds *DiskStorage) SetCookies(u *url.URL, cookies string) {
	//glog.Infof("disk storage run SetCookies, %v, %s", u, cookies)
	if err := os.WriteFile(ds.fileName, []byte(cookies), 0644); err != nil {
		log.Fatal(err)
	}
}
