package engine

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/model"
	"os"
	"strconv"
	"strings"
)

func (e *Engine) RunTradeData() {
	glog.Info("开始运行商场角色交易数据爬取")
	account := e.config.Accounts[0]
	if err := e.login(account); err != nil {
		glog.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.trade(account); err != nil {
		glog.Errorf("账号[%s]查询商场数据错误, 原因: %v", account.Username, err)
		return
	}
}

func (e *Engine) trade(account config.Account) error {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	var tradeResp model.TradeResp
	c.OnResponse(func(response *colly.Response) {
		respBody := response.Body
		err := json.Unmarshal(respBody, &tradeResp)
		if err != nil {
			glog.Errorf("账号[%s]商场角色交易数据解码Json错误, 返回内容: %s", name, string(respBody))
			return
		}
	})
	searchTradeData := map[string]string{
		"update":         "page",
		"timeout":        "false",
		"hovering":       "false",
		"tradehandler":   "",
		"service":        "charactertrade",
		"currency":       "coins",
		"realm":          "7",
		"character":      "",
		"currentmenu":    "-1",
		"currentsubmenu": "-1",
		"class":          "-1",
		"purchasetype":   "0",
		"purchasevalue":  "0",
		"page":           "0",
		"tradetab":       "",
		"selltab":        "",
		"method":         "load",
		"do":             "search",
	}
	/*
		searchTradeData := map[string]string{
			"update":    "page",
			"method":    "load",
			"realm":     "7",
			"character": "",
			"currency":  "coins",
			"service":   "charactertrade",
		}
	*/
	err := c.Post(constant.TradeUrl, searchTradeData)
	/*
		buf, _ := os.ReadFile("tradeResp.html")
	*/
	//doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(buf)))
	if tradeResp.Content == nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(tradeResp.Content[0]))
	if err != nil {
		return err
	}
	var trades []*model.TradeInfo
	var cc string
	doc.Find("tr[class!=static]").Each(func(i int, s *goquery.Selection) {
		ti := &model.TradeInfo{}
		a := s.Find("td[class^=name] > a")
		if url, exists := a.Attr("href"); exists {
			ti.ArmoryUrl = url
			ti.Name = a.Text()
		}
		ti.CharDesc = s.Find("td[class^=name] > div").Text()
		//ti.CharDesc = strings.Replace(s.Find("td[class^=name] > div").Text(), " ", "", -1)
		ti.Coins, _ = strconv.Atoi(s.Find("td[class=costValues] > span").Text())
		trades = append(trades, ti)
		cc = cc + "\n" + ti.Name + "\n" + ti.ArmoryUrl + "\n" + strconv.Itoa(ti.Coins) + ti.CharDesc
	})
	os.WriteFile("tradeInfo", []byte(cc), 0644)
	glog.Infof("商场角色交易数据写入成功, %d", len(cc))
	return err
}
