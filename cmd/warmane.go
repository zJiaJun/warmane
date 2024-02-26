package main

import (
	"flag"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/engine"
)

func init() {
	_ = flag.Set("log_dir", "./logs")
	flag.Parse()
}

func main() {
	glog.Info("Main engine start")
	e := engine.New()
	e.RunDailyPoints()
}
