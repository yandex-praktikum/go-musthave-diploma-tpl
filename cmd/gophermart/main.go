package main

import (
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/server"
)

func main() {
	logger.InitLogger()
	go server.StartServer()
	select {}
}
