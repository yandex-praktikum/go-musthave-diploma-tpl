// Package server предоставляет HTTP-сервер для приложения.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Config конфигурация HTTP сервера
type Config struct {
	Address         string        // Адрес и порт запуска
	ReadTimeout     time.Duration // Таймаут чтения запроса
	WriteTimeout    time.Duration // Таймаут записи ответа
	IdleTimeout     time.Duration // Таймаут keep-alive соединений
	ShutdownTimeout time.Duration // Таймаут graceful shutdown
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Address:         ":8080",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Server HTTP сервер
type Server struct {
	config     Config
	httpServer *http.Server
}

// New создаёт новый HTTP сервер
func New(config Config, handler http.Handler) *Server {
	return &Server{
		config: config,
		httpServer: &http.Server{
			Addr:         config.Address,
			Handler:      handler,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}
}

// Run запускает сервер и блокирует выполнение
func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
