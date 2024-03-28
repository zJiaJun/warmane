package main

import (
	"flag"
	"gitub.com/zJiajun/warmane/engine"
	"gitub.com/zJiajun/warmane/logger"
	"os"
)

var goflag = flag.NewFlagSet("warmane", flag.ExitOnError)
var (
	config string
	points bool
	trade  bool
)

func init() {
	goflag.StringVar(&config, "c", "config.yml", "Configuration file")
	goflag.BoolVar(&points, "p", false, "Run daily collect points")
	goflag.BoolVar(&trade, "t", true, "Run scraper trade data")
	goflag.Parse(os.Args[1:])
}

func main() {
	logger.Info("Main engine start")
	e := engine.New(config)
	if points {
		e.RunDailyPoints()
	}
	if trade {
		e.RunTradeData()
	}
	e.KeepSession()
	logger.Info("Main engine finish")
}
