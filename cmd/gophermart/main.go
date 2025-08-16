package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/config"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/logger"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/server"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/storage"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Подключаемся к базе данных
	dbStorage, err := storage.NewDatabaseStorage(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbStorage.Close()

	// Генерируем секретный ключ для JWT
	jwtSecret, err := services.GenerateSecret()
	if err != nil {
		log.Fatal("Failed to generate JWT secret", zap.Error(err))
	}

	// Создаем сервисы
	authService := services.NewAuthService(jwtSecret)
	accrualService := services.NewAccrualService(cfg.AccrualSystemAddress)

	// Создаем роутер
	router := server.NewRouter(dbStorage, authService, accrualService, log)

	orderProcessInterval, err := cfg.GetOrderProcessInterval()
	if err != nil {
		log.Fatal("Failed to parse order process interval", zap.Error(err))
	}

	// Создаем процессор заказов
	orderProcessor := server.NewOrderProcessor(dbStorage, accrualService, orderProcessInterval, cfg.WorkerCount, log)
	orderProcessor.Start()
	defer orderProcessor.Stop()

	// HTTP сервер
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router.GetRouter(),
	}

	// Обработкой сигналов для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Канал для передачи ошибок сервера
	serverErrors := make(chan error, 1)

	// Запускаем сервер в горутине
	go func() {
		log.Info("Starting server", zap.String("address", cfg.RunAddress))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Ждем либо сигнала завершения, либо ошибки сервера
	select {
	case <-ctx.Done():
		log.Info("Shutting down server...")
	case err := <-serverErrors:
		log.Error("Server error", zap.Error(err))
		log.Info("Shutting down server...")
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("Server stopped")
}
