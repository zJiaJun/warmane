package engine

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"gitub.com/zJiajun/warmane/config"
	"gitub.com/zJiajun/warmane/constant"
	"gitub.com/zJiajun/warmane/logger"
	"gitub.com/zJiajun/warmane/model"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

const (
	FrostwolfRealm = "4"
	LordaeronRealm = "6"
	IcecrownRealm  = "7"
	BlackrockRealm = "10"
	OnyxiaRealm    = "14"
)

func (e *Engine) RunTradeData() {
	logger.Info("开始运行商场角色交易数据爬取")
	account := e.config.Accounts[0]
	if err := e.login(account); err != nil {
		logger.Errorf("账号[%s]登录错误, 原因: %v", account.Username, err)
		return
	}
	if err := e.trade(account); err != nil {
		logger.Errorf("账号[%s]查询商场数据错误, 原因: %v", account.Username, err)
		return
	}
}

func (e *Engine) trade(account config.Account) error {
	name := account.Username
	c := e.getScraper(name).CloneCollector()
	e.getScraper(name).SetRequestHeaders(c)
	e.getScraper(name).DecodeResponse(c)
	var tradeResp struct {
		Content []string `json:"content"`
	}
	c.OnResponse(func(response *colly.Response) {
		respBody := response.Body
		if err := json.Unmarshal(respBody, &tradeResp); err != nil {
			logger.Errorf("账号[%s]商场角色交易数据解码Json错误, 返回内容: %s", name, string(respBody))
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
		"realm":          IcecrownRealm,
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

	err := c.Post(constant.TradeUrl, searchTradeData)

	if tradeResp.Content == nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(tradeResp.Content[0]))
	if err != nil {
		return err
	}
	var trades []*model.TradeInfo
	doc.Find("tr[class!=static]").Each(func(i int, s *goquery.Selection) {
		ti := &model.TradeInfo{}
		ti.Realm = IcecrownRealm
		a := s.Find("td[class^=name] > a")
		if url, exists := a.Attr("href"); exists {
			ti.ArmoryUrl = url
			ti.Name = a.Text()
		}
		/*
			CHARACTER INFORMATION
			Faction: Alliance
			Race: Draenei (male)
			Class: Shaman
			Level: 80
			Achievement Points: 1855
			Gold: 5
			Played time: 3 weeks
			Arena Points: 9
			Honor Points: 1301
			Inventory is not included, only the character's currently equipped items will be present upon purchase                                        Emblems are not included
		*/
		cd := s.Find("td[class^=name] > div").Text()
		var b strings.Builder
		for _, v := range strings.Split(cd, "\n") {
			tv := strings.TrimSpace(v)
			if tv == "" {
				continue
			}
			b.WriteString(tv + "\n")
		}
		ti.CharDesc = b.String()
		ti.Coins, _ = strconv.Atoi(s.Find("td[class=costValues] > span").Text())
		trades = append(trades, ti)
	})
	r := e.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "realm"}, {Name: "name"}},
		UpdateAll: true,
	}).Create(trades)
	logger.Infof("商场角色交易数据写入成功, %d", r.RowsAffected)
	return err
}

func (e *Engine) characterDetail() error {
	return nil
}
