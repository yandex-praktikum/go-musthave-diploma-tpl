package service

import (
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
)

type Service struct {
	db   Storage
	logs *logger.Logger
}

func NewService(db Storage, logger *logger.Logger) *Service {
	return &Service{
		db,
		logger,
	}
}
