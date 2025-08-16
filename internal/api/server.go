package api

import (
	"gorm.io/gorm"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/storage"
)

type Server struct {
	Config *config.AppConfig
	DB     *gorm.DB
}

func NewServer(config *config.AppConfig) *Server {

	return &Server{
		Config: config,
		DB:     storage.InitDB(*config),
	}
}
