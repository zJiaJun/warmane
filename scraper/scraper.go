package scraper

import (
	"github.com/gocolly/colly/v2"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/logger"
	"gitub.com/zJiajun/warmane/scraper/internal/decode"
	"gitub.com/zJiajun/warmane/scraper/internal/extensions"
	"gitub.com/zJiajun/warmane/scraper/storage"
	"gorm.io/gorm"
	"time"
)

type Scraper struct {
	c         *colly.Collector
	csrfToken string
}

func newScraper(name string, db *gorm.DB) *Scraper {
	s := &Scraper{
		c: colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.UserAgent(extensions.RandomUserAgent()),
			colly.IgnoreRobotsTxt(),
			//colly.Debugger(&debug.LogDebugger{}),
		),
	}
	s.c.SetRequestTimeout(10 * time.Second)
	if err := s.c.Limit(&colly.LimitRule{
		DomainGlob:  "*warmane.*",
		Parallelism: 1,
		Delay:       3 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	}); err != nil {
		panic(err)
	}
	/*
		if err := s.c.SetStorage(storage.NewDiskStorage(name)); err != nil {
			panic(err)
		}
	*/
	if err := s.c.SetStorage(storage.NewSqliteStorage(name, db)); err != nil {
		panic(err)
	}
	return s
}

func (s *Scraper) SetRequestHeaders(c *colly.Collector) {
	c.OnRequest(func(request *colly.Request) {
		request.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		request.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		request.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		request.Headers.Set("Cache-Control", "no-cache")
		request.Headers.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		request.Headers.Set("Origin", constant.BaseUrl)
		request.Headers.Set("Pragma", "no-cache")
		request.Headers.Set("Referer", constant.LoginUrl)
		if s.csrfToken == "" {
			e := c.Clone()
			e.OnHTML("meta[name='csrf-token']", func(element *colly.HTMLElement) {
				s.csrfToken = element.Attr("content")
				logger.Infof("查询获取warmane网站的csrfToken成功: %s", s.csrfToken)
			})
			_ = e.Visit(constant.BaseUrl)
		}
		request.Headers.Set("X-Csrf-Token", s.csrfToken)
		request.Headers.Set("X-Requested-With", "XMLHttpRequest")
	})
}

func (s *Scraper) DecodeResponse(c *colly.Collector) {
	c.OnResponse(func(response *colly.Response) {
		encoding := response.Headers.Get("Content-Encoding")
		if encoding == "" {
			return
		}
		decodeResp, err := decode.ResponseBody(encoding, response.Body)
		if err != nil {
			logger.Errorf("onResponse decode [%s] response error, %v", response.Request.URL, err)
			return
		}
		response.Body = decodeResp
	})
}

func (s *Scraper) CloneCollector() *colly.Collector {
	return s.c.Clone()
}
