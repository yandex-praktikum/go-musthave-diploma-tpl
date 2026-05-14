package config

import (
	"github.com/ilyakaznacheev/cleanenv"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/handler"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/service"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/worker"
)

type Config struct {
	Handler handler.Config `yaml:"Handler"`

	Service service.Config `yaml:"Service"`

	Worker worker.Config `yaml:"Worker"`

	Addr     string `yaml:"Addr"`
	LogLevel string `yaml:"LogLevel"`

	DatabaseDSN string `yaml:"DatabaseDSN"`
}

// Чтение конфига из файла
func NewConfig(configPath string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
