package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr            string
	ResultAddr      string
	FileStoragePath string
	DatabaseDsn     string
}

func NewConfig() Config {
	return Config{}
}

func (cfg *Config) InitConfig() Config {
	cfg.argsConfigurator()
	cfg.envConfigurator()

	return Config{
		Addr:            cfg.Addr,
		ResultAddr:      cfg.ResultAddr,
		FileStoragePath: cfg.FileStoragePath,
		DatabaseDsn:     cfg.DatabaseDsn,
	}
}

func (cfg *Config) argsConfigurator() {
	addr := flag.String("a", "localhost:8080", "server ip:port")
	resultAddr := flag.String("b", "http://localhost:8080", "server result ip:port")
	fileStoragePath := flag.String("f", "storageFile", "file storage path")
	databaseDsn := flag.String("d", "", "database connection string")

	flag.Parse()

	cfg.Addr = *addr
	cfg.ResultAddr = *resultAddr
	cfg.FileStoragePath = *fileStoragePath
	cfg.DatabaseDsn = *databaseDsn
}

func (cfg *Config) envConfigurator() {
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envResultAddr := os.Getenv("BASE_URL"); envResultAddr != "" {
		cfg.ResultAddr = envResultAddr
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		cfg.DatabaseDsn = envDatabaseDsn
	}
}
