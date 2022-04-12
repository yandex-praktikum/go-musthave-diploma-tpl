package cfg

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr        string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`                                   // Адрес запуска HTTP-сервера
	CryptoKey      string `env:"CRYPTO_KEY" envDefault:"secret_123456789"`                                  // secret word to encrypt/decrypt JWT for cookies
	DatabaseURI    string `env:"DATABASE_URI" envDefault:"postgresql://localhost:5432/yandex_practicum_db"` // Строка с адресом подключения к БД
	AccrualSysAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`                                                    // Адрес системы расчёта начислений «http://server:port»
}

var Envs Config

type ContextKey string

func GetEnvs() error {
	flag.StringVar(&Envs.RunAddr, "a", "http://localhost:8080", "RUN_ADDRESS to listen on")
	flag.StringVar(&Envs.DatabaseURI, "d", "postgresql://localhost:5432/yandex_practicum_db", "DATABASE_DSN. Address for connection to DB")
	flag.StringVar(&Envs.AccrualSysAddr, "r", "", "ACCRUAL_SYSTEM_ADDRESS ")

	if err := env.Parse(&Envs); err != nil {
		return err
	}

	flag.Parse()

	return nil
}
