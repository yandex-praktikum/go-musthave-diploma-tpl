package main

import (
	"context"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/handlers"
	"go-musthave-diploma-tpl/internal/logger"
	"go-musthave-diploma-tpl/internal/server"
	"go-musthave-diploma-tpl/internal/services"
	"go-musthave-diploma-tpl/internal/storage"
	"go-musthave-diploma-tpl/internal/worker"
)

func main() {
	// Инициализируем логгер
	logConfig := logger.DefaultConfig()
	if os.Getenv("ENV") == "production" {
		logConfig.Format = "json"
		logConfig.Level = logger.LevelInfo
		logConfig.AddSource = false
	}
	log := logger.New(logConfig)
	logger.SetDefault(log)

	logger.Info("Starting Gophermart API")

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		logger.ErrorWithStack("Failed to load config", err)
		os.Exit(1)
	}

	logger.Info("Config loaded",
		"run_address", cfg.RunAddress,
		"accrual_address", cfg.AccrualSystemAddress,
	)

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем хранилище
	store, err := storage.New(ctx, cfg.DatabaseURI)
	if err != nil {
		logger.ErrorWithStack("Failed to initialize storage", err)
		os.Exit(1)
	}
	defer store.Close()

	logger.Info("Storage initialized successfully")

	// Инициализируем сервисы
	tokenManager := auth.NewTokenManager("gophermart-secret-key", time.Hour*24*30)
	userService := services.NewUserService(store.Users, tokenManager)
	orderService := services.NewOrderService(store.Orders, store.Balance)

	// Инициализируем клиент accrual системы
	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress, time.Second*30)

	// Инициализируем воркер обработки заказов
	workerConfig := worker.DefaultConfig()
	orderWorker := worker.New(workerConfig, orderService, accrualClient)

	// Запускаем воркер
	orderWorker.Run(ctx)
	logger.Info("Order worker started",
		slog.Duration("poll_interval", workerConfig.PollInterval),
		slog.Int("batch_limit", workerConfig.BatchLimit),
	)

	// Инициализируем хендлеры
	authHandler := handlers.NewAuthHandler(userService)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Создаем роутер
	router := server.Router()

	// Публичные роуты
	router.Post("/api/user/register", authHandler.Register)
	router.Post("/api/user/login", authHandler.Login)

	// Защищенные роуты
	authMiddleware := auth.NewAuthMiddleware(tokenManager)
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)

		// Заказы
		r.Post("/api/user/orders", orderHandler.UploadOrder)
		r.Get("/api/user/orders", orderHandler.GetOrders)

		// Баланс
		r.Get("/api/user/balance", orderHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", orderHandler.Withdraw)
		r.Get("/api/user/withdrawals", orderHandler.GetWithdrawals)
	})

	// Health check
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gophermart API"))
	})

	// Создаем HTTP сервер
	serverConfig := server.DefaultConfig()
	serverConfig.Address = cfg.RunAddress

	httpServer := server.New(serverConfig, router)

	// Канал для сигналов завершения
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		logger.Info("Starting server", "address", cfg.RunAddress)
		if err := httpServer.Run(); err != nil {
			logger.ErrorWithStack("Server error", err)
		}
	}()

	// Ждём сигнал завершения
	<-shutdown
	logger.Info("Shutting down...")

	// Graceful shutdown с таймаутом
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем воркер
	logger.Info("Stopping order worker...")
	if err := orderWorker.Shutdown(shutdownCtx); err != nil {
		logger.Error("Worker shutdown error", "error", err)
	}

	// Останавливаем сервер
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Graceful shutdown error", "error", err)
	}

	logger.Info("Server stopped")
}
