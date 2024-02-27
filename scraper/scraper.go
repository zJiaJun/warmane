package scraper

import (
	"github.com/gocolly/colly/v2"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/scraper/internal/decode"
	"gitub.com/zJiajun/warmane/scraper/internal/extensions"
	"gitub.com/zJiajun/warmane/scraper/internal/storage"
	"time"
)

type Scraper struct {
	c *colly.Collector
}

func NewScraper(name string) *Scraper {
	s := &Scraper{
		c: colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.UserAgent(extensions.RandomUserAgent()),
			colly.IgnoreRobotsTxt(),
		),
	}
	s.c.SetRequestTimeout(60 * time.Second)
	if err := s.c.SetStorage(storage.NewDiskStorage(name)); err != nil {
		panic(err)
	}
	return s
}

func (s *Scraper) SetRequestHeaders(c *colly.Collector, csrfToken string) {
	c.OnRequest(func(request *colly.Request) {
		request.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		request.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		request.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		request.Headers.Set("Cache-Control", "no-cache")
		request.Headers.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
		request.Headers.Set("Origin", config.BaseUrl)
		request.Headers.Set("Pragma", "no-cache")
		request.Headers.Set("Referer", config.LoginUrl)
		if csrfToken != "" {
			request.Headers.Set("X-Csrf-Token", csrfToken)
		}
		request.Headers.Set("X-Requested-With", "XMLHttpRequest")
	})
}

func (s *Scraper) DecodeResponse(c *colly.Collector) {
	c.OnResponse(func(response *colly.Response) {
		//glog.Infof("onResponse [%s], statusCode:[%d], size:[%d]", response.Request.URL, response.StatusCode, len(response.Body))
		encoding := response.Headers.Get("Content-Encoding")
		if encoding == "" {
			return
		}
		decodeResp, err := decode.ResponseBody(encoding, response.Body)
		if err != nil {
			glog.Errorf("onResponse decode [%s] response error, %v", response.Request.URL, err)
			return
		}
		response.Body = decodeResp
		//glog.Infof("onResponse decode [%s] response success", response.Request.URL)
	})
}

func (s *Scraper) CloneCollector() *colly.Collector {
	return s.c.Clone()
}