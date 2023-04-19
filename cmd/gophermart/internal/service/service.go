package service

import (
	"github.com/RedWood011/cmd/gophermart/internal/config"
	"golang.org/x/exp/slog"
)

type Service struct {
	storage Storage
	cfg     *config.Config
	logger  *slog.Logger
}

func NewService(storage Storage, cfg *config.Config, logger *slog.Logger) *Service {
	return &Service{
		storage: storage,
		cfg:     cfg,
		logger:  logger,
	}
}
