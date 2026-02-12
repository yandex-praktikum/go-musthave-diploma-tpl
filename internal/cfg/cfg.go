package cfg

import (
	"context"
	"flag"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

var (
	once   sync.Once
	config Config
)

type Config struct {
	Address          string `env:"RUN_ADDRESS"`
	DbDSN            string `env:"DATABASE_URI"`
	AccrualSystemUrl string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func GetConfig() *Config {
	once.Do(func() {
		var tempConfig Config

		if err := envconfig.Process(context.Background(), &tempConfig); err != nil {
			logger.Error("Failed to load config from env", zap.Error(err))
		}

		if tempConfig.Address == "" {
			flag.StringVar(&tempConfig.Address, "a", "localhost:7000", "адрес и порт запуска сервиса")
		}

		if tempConfig.DbDSN == "" {
			flag.StringVar(&tempConfig.DbDSN, "d", "", "адрес подключения к базе данных") // TODO
		}

		if tempConfig.AccrualSystemUrl == "" {
			flag.StringVar(&tempConfig.AccrualSystemUrl, "r", "", "адрес системы расчёта начислений") // TODO
		}

		flag.Parse()

		config = tempConfig
	})

	return &config
}
