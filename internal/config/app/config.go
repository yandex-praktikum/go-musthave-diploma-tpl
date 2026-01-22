package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Port           string
	AccrualPath    string
	DNS            string
	ParamGetStatus int
	SecretKey      string
	ParamTimeOut   time.Duration // тай
}

// NewConfig - создание конфигурации приложения
func NewConfig() *Config {
	cfg := Config{}
	var paramTimeOut int
	// обязателльные параметры
	flag.StringVar(&cfg.Port, "a", getDef("RUN_ADDRESS", "localhost:8080"), "порт сервиса")
	flag.StringVar(&cfg.DNS, "d", getDef("DATABASE_URI", "postgres://postgres:12345678@localhost:5432/market?sslmode=disable"), "cтрока с адресом подключения к БД")
	flag.StringVar(&cfg.AccrualPath, "r", getDef("ACCRUAL_SYSTEM_ADDRESS", ""), "путь к blackBox")

	// кастомные
	flag.StringVar(&cfg.SecretKey, "k", "tort-secret-key", "ключ")
	flag.IntVar(&cfg.ParamGetStatus, "t", 20, "частота запуска очистки от помеченных на удаление URL")
	flag.IntVar(&paramTimeOut, "s", 20, "таймаут подключения к внешнему сервису(в секундах)")
	flag.Parse()

	cfg.ParamTimeOut = time.Duration(paramTimeOut) * time.Second

	//if runAddr, exists := os.LookupEnv("RUN_ADDRESS"); exists && runAddr != "" {
	//	cfg.Port = runAddr
	//}
	//
	//if accrualPath, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); exists && accrualPath != "" {
	//	cfg.AccrualPath = accrualPath
	//}
	//if db, exists := os.LookupEnv("DATABASE_URI"); exists && db != "" {
	//	cfg.DNS = db
	//}

	return &cfg
}

func getDef(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
