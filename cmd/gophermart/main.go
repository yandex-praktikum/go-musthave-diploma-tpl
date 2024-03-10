package main

import (
	"github.com/A-Kuklin/gophermart/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.LoadConfig()
	logrus.SetLevel(cfg.LogLevel)
}
