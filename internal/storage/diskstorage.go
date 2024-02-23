package storage

import (
	"fmt"
	"github.com/gocolly/colly/v2/storage"
	"github.com/golang/glog"
	"log"
	"net/url"
	"os"
)

type DiskStorage struct {
	fileName        string
	inMemoryStorage *storage.InMemoryStorage
}

func NewDiskStorage() *DiskStorage {
	return &DiskStorage{
		fileName:        ".cookies",
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
	glog.Infof("disk storage run cookies, %v", u)
	filePath := fmt.Sprintf("%s"+ds.fileName, u.Hostname())
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (ds *DiskStorage) SetCookies(u *url.URL, cookies string) {
	glog.Infof("disk storage run SetCookies, %v, %s", u, cookies)
	filePath := fmt.Sprintf("%s"+ds.fileName, u.Hostname())
	if err := os.WriteFile(filePath, []byte(cookies), 0644); err != nil {
		log.Fatal(err)
	}
}
