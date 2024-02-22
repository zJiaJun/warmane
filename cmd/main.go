package main

import (
	"flag"
	"github.com/golang/glog"
	"gitub.com/zJiajun/warmane/warmane"
)

func init() {
	_ = flag.Set("log_dir", "./")
	flag.Parse()
}

func main() {
	glog.Info("Main engine start")
	w := warmane.New()
	w.RunDailyPoints()
}
