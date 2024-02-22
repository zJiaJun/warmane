package scraper

import (
	"github.com/gocolly/colly/v2"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/internal/config"
	"gitub.com/zJiajun/warmane/internal/decode"
)

type Scraper struct {
	c *colly.Collector
}

func NewScraper() *Scraper {
	return &Scraper{
		c: colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.UserAgent(RandomUserAgent()),
			colly.IgnoreRobotsTxt(),
		),
	}
}

func (s *Scraper) SetRequestHeaders(c *colly.Collector, csrfToken string) {
	c.OnRequest(func(request *colly.Request) {
		request.Headers.Set("Origin", config.BaseUrl)
		request.Headers.Set("Referer", config.LoginUrl)
		request.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		request.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		request.Headers.Set("X-Csrf-Token", csrfToken)
		request.Headers.Set("X-Requested-With", "XMLHttpRequest")
		request.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7")
		request.Headers.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	})
}

func (s *Scraper) DecodeResponse(c *colly.Collector) {
	c.OnResponse(func(response *colly.Response) {
		glog.Infof("onResponse [%s], statusCode:[%d], size:[%d]", response.Request.URL, response.StatusCode, len(response.Body))
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
		glog.Infof("onResponse decode [%s] response success", response.Request.URL)
	})
}

func (s *Scraper) CloneCollector() *colly.Collector {
	return s.c.Clone()
}