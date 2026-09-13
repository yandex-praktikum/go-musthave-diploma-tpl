// Command gophermart запускает HTTP-сервер сервиса лояльности Gophermart:
// загружает конфигурацию, подключается к PostgreSQL, применяет миграции,
// инициализирует репозитории, сервисы и обработчики, запускает фоновый
// воркер расчёта вознаграждений и обеспечивает корректное завершение
// по сигналам SIGINT/SIGTERM.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/config"
	"github.com/b602op/go-musthave-diploma-tpl/internal/db"
	"github.com/b602op/go-musthave-diploma-tpl/internal/handlers"
	"github.com/b602op/go-musthave-diploma-tpl/internal/middleware"
	"github.com/b602op/go-musthave-diploma-tpl/internal/repository"
	"github.com/b602op/go-musthave-diploma-tpl/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Println("=== Starting Gophermart Service ===")

	// 1. Создаём контекст, который отменяется по SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 2. Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config loaded: RUN_ADDRESS=%s, ACCRUAL_ADDRESS=%s", cfg.RunAddress, cfg.AccrualSystemAddress)

	// 3. Подключаемся к БД
	dbConn, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbConn.Close()

	// Проверяем соединение
	log.Println("Connecting to database...")
	if err := dbConn.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connected successfully")

	// 4. Применяем миграции автоматически
	if err := db.RunMigrations(dbConn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 5. Создаем репозитории
	log.Println("Initializing repositories...")
	userRepo := repository.NewUserRepository(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	balanceRepo := repository.NewBalanceRepository(dbConn)
	withdrawalRepo := repository.NewWithdrawalRepository(dbConn)
	log.Println("Repositories initialized")

	// 6. Создаем клиент для accrual
	log.Printf("Initializing accrual client: %s", cfg.AccrualSystemAddress)
	accrualClient := service.NewAccrualClient(cfg.AccrualSystemAddress)
	log.Println("Accrual client initialized")

	// 7. Создаем сервисы
	log.Println("Initializing services...")
	authService := service.NewAuthService(userRepo)
	orderService := service.NewOrderService(orderRepo, balanceRepo, accrualClient)
	balanceService := service.NewBalanceService(balanceRepo, withdrawalRepo, dbConn)
	log.Println("Services initialized")

	// 8. Запускаем фоновый воркер для обработки заказов.
	worker := service.NewWorker(orderService)
	go worker.Start(ctx)
	log.Println("Worker started")

	// 9. Создаем middleware с секретным ключом
	authMiddleware := middleware.NewAuthMiddleware(cfg.SecretKey)
	log.Println("Auth middleware initialized")

	// 10. Создаем обработчики
	log.Println("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(authService, authMiddleware)
	orderHandler := handlers.NewOrderHandler(orderService)
	balanceHandler := handlers.NewBalanceHandler(balanceService)
	log.Println("Handlers initialized")

	// 11. Настраиваем роутер
	log.Println("Setting up router...")
	router := setupRouter(authHandler, orderHandler, balanceHandler, authMiddleware)
	log.Println("Router configured")

	// 12. Создаем HTTP сервер
	server := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 13. Запускаем сервер в горутине
	go func() {
		log.Printf("Starting server on %s", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// 14. Ждём сигнала от ОС (или отмены родительского контекста)
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Восстанавливаем поведение по умолчанию для сигналов,
	// чтобы повторный Ctrl+C прервал процесс немедленно.
	stop()

	// 15. Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server shutdown failed:", err)
	}

	log.Println("Server stopped")
}

// setupRouter настраивает все маршруты API: публичные (регистрация,
// вход) и защищённые middleware аутентификации (заказы, баланс, списания).
// Оборачивает роутер middleware логирования и восстановления после паник.
func setupRouter(
	authHandler *handlers.AuthHandler,
	orderHandler *handlers.OrderHandler,
	balanceHandler *handlers.BalanceHandler,
	authMiddleware *middleware.AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	// Публичные маршруты (без авторизации)
	mux.HandleFunc("POST /api/user/register", authHandler.Register)
	mux.HandleFunc("POST /api/user/login", authHandler.Login)

	// Защищенные маршруты (с авторизацией)
	mux.Handle("POST /api/user/orders", authMiddleware.Auth(http.HandlerFunc(orderHandler.Upload)))
	mux.Handle("GET /api/user/orders", authMiddleware.Auth(http.HandlerFunc(orderHandler.List)))
	mux.Handle("GET /api/user/balance", authMiddleware.Auth(http.HandlerFunc(balanceHandler.GetBalance)))
	mux.Handle("POST /api/user/balance/withdraw", authMiddleware.Auth(http.HandlerFunc(balanceHandler.Withdraw)))
	mux.Handle("GET /api/user/withdrawals", authMiddleware.Auth(http.HandlerFunc(balanceHandler.GetWithdrawals)))

	// Добавляем middleware для логирования и восстановления после паник
	return middleware.Logging(middleware.Recovery(mux))
}
