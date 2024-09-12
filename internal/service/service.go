package service

import (
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
)

type Service struct {
	db   StorageServ
	logs *logger.Logger
}

func NewService(db StorageServ, logger *logger.Logger) *Service {
	return &Service{
		db,
		logger,
	}
}
