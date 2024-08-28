package engine

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/zJiajun/warmane/common"
	"github.com/zJiajun/warmane/constant"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/model"
	"github.com/zJiajun/warmane/model/table"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

var (
	frostwolfRealm = &common.Pair[string, string]{Left: "4", Right: "Frostwolf"}
	lordaeronRealm = &common.Pair[string, string]{Left: "6", Right: "Lordaeron"}
	icecrownRealm  = &common.Pair[string, string]{Left: "7", Right: "Icecrown"}
	blackrockRealm = &common.Pair[string, string]{Left: "10", Right: "Blackrock"}
	onyxiaRealm    = &common.Pair[string, string]{Left: "14", Right: "Onyxia"}
)

func (e *Engine) trade(account *table.Account) error {
	trades, err := e.fetchTradeData(account)
	if err != nil {
		return err
	}
	err = e.storeTradeData(account.AccountName, trades)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) fetchTradeData(account *table.Account) ([]*table.TradeInfo, error) {
	name := account.AccountName
	c := e.getScraper(name).CloneCollector()
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
		"realm":          icecrownRealm.Left,
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
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(tradeResp.Content[0]))
	if err != nil {
		return nil, err
	}
	var trades []*table.TradeInfo
	doc.Find("tr[class!=static]").Each(func(i int, s *goquery.Selection) {
		ti := &table.TradeInfo{
			BasicCharacter: &model.BasicCharacter{Realm: icecrownRealm.Right},
		}
		a := s.Find("td[class^=name] > a")
		if url, exists := a.Attr("href"); exists {
			ti.ArmoryUrl = url
			ti.BasicCharacter.Name = a.Text()
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
		ti.Coins, _ = strconv.Atoi(s.Find("td[class=costValues] > span").Text())
		cd := s.Find("td[class^=name] > div").Text()
		var b strings.Builder
		var inventoryIncluded int
		for _, v := range strings.Split(cd, "\n") {
			tv := strings.TrimSpace(v)
			if tv == "" {
				continue
			}
			if strings.Contains(tv, "Inventory is not included") {
				inventoryIncluded = 0
			} else {
				inventoryIncluded = 1
			}
			b.WriteString(tv + "\n")
		}
		ti.CharDesc = b.String()
		ti.InventoryIncluded = inventoryIncluded
		trades = append(trades, ti)
	})
	return trades, nil
}

func (e *Engine) storeTradeData(name string, trades []*table.TradeInfo) error {
	r := e.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "realm"}, {Name: "name"}},
		UpdateAll: true,
	}).Create(trades)
	if r.Error != nil {
		return r.Error
	}
	logger.Infof("商场角色交易数据写入成功, %d", r.RowsAffected)
	return nil
}

func (e *Engine) fillCharacterDetail(name string, ti *table.TradeInfo) error {
	armoryApiUrl := strings.ReplaceAll(ti.ArmoryUrl, "armory.warmane.com", "armory.warmane.com/api")
	c := e.getScraper(name).CloneCollector()
	c.OnResponse(func(response *colly.Response) {
		if err := json.Unmarshal(response.Body, &ti.BasicCharacter); err != nil {
			return
		}
	})
	err := c.Visit(armoryApiUrl)
	return err
}
