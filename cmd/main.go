package main

import (
	"flag"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/engine"
)

var config string

func init() {
	_ = flag.Set("log_dir", "./logs")
	flag.StringVar(&config, "config", "config.yml", "configuration file")
	flag.Parse()
}

func main() {
	glog.Info("Main engine start")
	defer glog.Flush()
	e := engine.New(config)
	e.RunDailyPoints()
	//e.RunTradeData()
	glog.Infof("Main engine finish")

}
