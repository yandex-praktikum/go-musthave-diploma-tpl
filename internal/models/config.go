package models

// Config содержит все параметры конфигурации приложения
type Config struct {
	Address     string
	Server      string
	Port        string
	DatabaseDSN string
	Key         string // ключ для шифрования
	RateLimit   int64  // количество потоков
}
