package main

import (
	"fmt"
	"github.com/issimo1/redis-issimo/tcpserver"
	"github.com/issimo1/redis-issimo/utils/config"
	"github.com/issimo1/redis-issimo/utils/logger"
	"log"
	"os"
)

func main() {
	config.Init()
	logger.Init()
	logger.Info("Starting issimo1 redis server...")
	tcp := tcpserver.NewTCPServer(tcpserver.TCPConfig{
		Addr: fmt.Sprintf("%s:%d", config.GlobalConfig.Bind, config.GlobalConfig.Port)}, "redis")
	if err := tcp.Start(); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
	tcp.Close()
}
