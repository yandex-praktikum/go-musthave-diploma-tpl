package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const schema = "http"

func LoadConfig(configPath string) (*AppConfig, error) {
	config := &AppConfig{}

	// Загрузка конфигурации из файла
	if data, err := os.ReadFile(configPath); err == nil {
		if err = yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
		}
	}

	// Объявление флагов
	var flagRunAddress, flagDataBaseURI, flagAccuralSystemAddress string
	flag.StringVar(&flagRunAddress, "a", "", "адрес и порт запуска сервиса")
	flag.StringVar(&flagDataBaseURI, "d", "", "строка подключения к БД")
	flag.StringVar(&flagAccuralSystemAddress, "r", "", "адрес системы расчёта начислений") // Исправлено: "-r" → "r"
	flag.Parse()

	// Установка значений с приоритетами: флаг > env > файл > значение по умолчанию
	setConfigValue(&config.RunAddress, flagRunAddress, "RUN_ADDRESS", "localhost:8080")
	setConfigValue(&config.DatabaseURI, flagDataBaseURI, "DATABASE_URI",
		"user=postgres password=postgres host=localhost port=5432 database=pgx_test sslmode=disable")
	setConfigValue(&config.AccrualSystemAddress, flagAccuralSystemAddress, "ACCRUAL_SYSTEM_ADDRESS", "")

	return config, nil
}

// Вспомогательная функция для установки значений с приоритетами
func setConfigValue(field *string, flagValue, envKey, defaultValue string) {
	if flagValue != "" {
		*field = flagValue
		return
	}

	if env := os.Getenv(envKey); env != "" {
		*field = env
		return
	}

	if *field == "" && defaultValue != "" {
		*field = defaultValue
	}
}
