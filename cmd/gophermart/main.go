package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Используем хранилища как реализации интерфейсов repository
	tokenManager := auth.NewTokenManager("gophermart-secret-key", time.Hour*24*30)
	userService := services.NewUserService(store.Users, tokenManager)
	orderService := services.NewOrderService(store.Orders, store.Balance)

	// Инициализируем клиент accrual системы с опциями
	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress,
		accrual.WithTimeout(time.Second*30),
	)

	// Инициализируем воркер обработки заказов с опциями
	orderWorker := worker.New(worker.NewConfig(
		worker.WithPollInterval(time.Second*1),
		worker.WithBatchLimit(10),
		worker.WithRetryDelay(time.Second*60),
	), orderService, accrualClient)

	// Запускаем воркер
	orderWorker.Run(ctx)
	logger.Info("Order worker started")

	// Инициализируем хендлеры
	authHandler := handlers.NewAuthHandler(userService)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Создаём роутер (ServeMux)
	mux := server.Router()

	// Middleware аутентификации
	authMiddleware := auth.NewAuthMiddleware(tokenManager)

	// Вспомогательная функция для применения middleware к хендлеру
	withAuth := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.Middleware(handler).ServeHTTP(w, r)
		}
	}

	// Публичные роуты
	mux.HandleFunc("POST /api/user/register", authHandler.Register)
	mux.HandleFunc("POST /api/user/login", authHandler.Login)

	// Защищённые роуты (с middleware аутентификации)
	mux.HandleFunc("POST /api/user/orders", withAuth(orderHandler.UploadOrder))
	mux.HandleFunc("GET /api/user/orders", withAuth(orderHandler.GetOrders))
	mux.HandleFunc("GET /api/user/balance", withAuth(orderHandler.GetBalance))
	mux.HandleFunc("POST /api/user/balance/withdraw", withAuth(orderHandler.Withdraw))
	mux.HandleFunc("GET /api/user/withdrawals", withAuth(orderHandler.GetWithdrawals))

	// Health check
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gophermart API"))
	})

	// Применяем глобальные middleware к роутеру
	handler := server.WithMiddleware(mux,
		server.RealIP,
		server.RequestID,
		server.Recovery,
		server.RequestLogger,
		server.Gzip,
		server.GzipRequest,
	)

	// Создаем HTTP сервер
	serverConfig := server.DefaultConfig()
	serverConfig.Address = cfg.RunAddress

	httpServer := server.New(serverConfig, handler)

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
