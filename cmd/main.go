package main

import (
	"flag"
	"github.com/zJiajun/warmane/engine"
	"github.com/zJiajun/warmane/logger"
	"github.com/zJiajun/warmane/route"
	"net"
	"os"
	"sync"
)

var goflag = flag.NewFlagSet("warmane", flag.ExitOnError)

var port string

func init() {
	goflag.StringVar(&port, "p", "8070", "Port to listen on")
	goflag.Parse(os.Args[1:])
}

var onceMap = make(map[int]sync.Once)

func main() {
	logger.Info("Main engine start")
	e := engine.New()
	r := route.New(e)
	addr := net.JoinHostPort("", port)
	if err := r.Run(addr); err != nil {
		logger.Fatal(err)
	}
}
