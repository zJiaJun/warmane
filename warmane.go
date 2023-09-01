package main

import (
	"flag"
	"github.com/golang/glog"
)

func init() {
	flag.Parse()
}
func main() {

	defer glog.Flush()
	glog.Info("this is info log")
	glog.Warning("this is waring log")
	glog.Error("this is error log")
	glog.Fatal("this is fatal log")
}
