package main

import (
	"flag"
	"github.com/zJiajun/warmane/engine"
	"github.com/zJiajun/warmane/logger"
	"os"
)

var goflag = flag.NewFlagSet("warmane", flag.ExitOnError)
var (
	config      string
	points      bool
	trade       bool
	keepSession bool
)

func init() {
	goflag.StringVar(&config, "c", "config.yml", "Configuration file")
	goflag.BoolVar(&points, "p", false, "Run daily collect points")
	goflag.BoolVar(&trade, "t", false, "Run scraper trade data")
	goflag.BoolVar(&keepSession, "k", false, "Run keep session job")
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
	if keepSession {
		e.KeepSession()
	}
	logger.Info("Main engine finish")
}
