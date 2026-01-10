package config

import (
	"flag"
	"os"
)

type Config struct {
	Port        string
	AccrualPath string
	DNS         string
	//ParamDelete int
	SecretKey string
}

// NewConfig - создание конфигурации приложения
func NewConfig() *Config {
	cfg := Config{}
	// обязателльные параметры
	flag.StringVar(&cfg.Port, "a", ":8080", "порт сервиса")
	flag.StringVar(&cfg.DNS, "d", "postgres://postgres:12345678@localhost:5432/myDB?sslmode=disable", "cтрока с адресом подключения к БД")
	flag.StringVar(&cfg.AccrualPath, "r", "default", "путь к blackBox")

	//flag.IntVar(&cfg.ParamDelete, "t", 20, "частота запуска очистки от помеченных на удаление URL")
	// кастомные
	flag.StringVar(&cfg.AccrualPath, "r", "default", "путь к blackBox")
	flag.StringVar(&cfg.SecretKey, "k", "tort-secret-key", "ключ")

	flag.Parse()

	if runAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists && runAddr != "" {
		cfg.Port = runAddr
	}

	if accrualPath, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); exists && accrualPath != "" {
		cfg.AccrualPath = accrualPath
	}
	if db, exists := os.LookupEnv("DATABASE_DSN"); exists && db != "" {
		cfg.DNS = db
	}

	return &cfg
}
