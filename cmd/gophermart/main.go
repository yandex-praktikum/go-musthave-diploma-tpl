package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/postgresql"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/config"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/delivery"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/delivery/handlers"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/usecase"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/accrualclient"

	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("Error syncing logger: %v\n", syncErr)
		}
	}()
	zap.ReplaceGlobals(logger)

	// Загрузка конфигурации
	cfg, errLoadConfig := loadConfig(logger)
	if errLoadConfig != nil {
		logger.Fatal("can't load config", zap.Error(errLoadConfig))
	}

	// Инициализация хранилищ
	userRepo, errInitDB := initDatabase(cfg)
	if errInitDB != nil {
		logger.Fatal("can't init database", zap.Error(errInitDB))
	}

	usecaseUser := usecase.NewUsecase(
		userRepo,
		cfg.SecretKey, // Секрет для JWT
		12,            // Стоимость хеширования (bcrypt cost)
		1*time.Hour,
		24*time.Hour)

	fmt.Println(usecaseUser.GetUserBalance(context.Background(), 1))
	orderUseCase := getOrderUseCase(userRepo)
	if err := orderUseCase.StartOrderProcessing(context.Background()); err != nil {
		log.Fatal("Failed to start order processing:", err)
	}
	/*
		orders := []string{
			"79927398713",
		}

		for _, order := range orders {
			// Для каждого заказа можно задать свой интервал
			interval := 10 * time.Second
			if order == "79927398713" {
				interval = 5 * time.Second // Чаще проверяем
			}

			if err := workerPool.Add(ctx, order, interval); err != nil {
				log.Printf("Failed to add %s: %v", order, err)
			}
		}

		go func() {
			for result := range workerPool.Results() {
				if result.Error != nil {
					log.Printf("❌ %s: %v", result.OrderNumber, result.Error)
					continue
				}

				if result.OrderInfo != nil {
					fmt.Printf("✅ %s: %s (accrual: %.2f)\n",
						result.OrderNumber,
						result.OrderInfo.Status,
						result.OrderInfo.Order)
				}
				if result.OrderInfo.Status == orderclient.StatusProcessed {
					if err := workerPool.Remove(result.OrderNumber); err != nil {
						log.Printf("Failed to remove order: %v", err)
					}
				}
			}
		}()

		go func() {
			time.Sleep(30 * time.Second)

			// Добавляем новый заказ
			if err := workerPool.Add(ctx, "79927398713", 15*time.Second); err != nil {
				log.Printf("Failed to add new order: %v", err)
			}

		}()

	*/

	// Создание обработчика URL
	handler := handlers.NewUserHandler(
		usecaseUser,
		orderUseCase,
		cfg.RunAddress,
		cfg.AccrualSystemAddress,
		repository.NewSessionStore(),
		logger)

	// Инициализация сервера
	server, errInitServer := initServer(cfg, handler)
	if errInitServer != nil {
		logger.Fatal("can't init server", zap.Error(errInitServer))
	}

	// Запуск сервера
	runServer(server)

}

// runServer управляет жизненным циклом HTTP-сервера.
// Запускает сервер в отдельной горутине и обрабатывает сигналы завершения работы.
// Параметры:
// - server: экземпляр HTTP-сервера для управления
// Алгоритм:
// - Запускает сервер в отдельном канале для обработки ошибок
// - Ожидает сигналов OS Interrupt или SIGTERM
// - Выполняет graceful shutdown с таймаутом 10 секунд
func runServer(server *delivery.Server) {
	serverErrChan := server.Start()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		zap.L().Info("Received shutdown signal")
	case err := <-serverErrChan:
		if err != nil {
			zap.L().Error("Server error", zap.Error(err))
		}
	}

	if err := server.Shutdown(10 * time.Second); err != nil {
		zap.L().Error("Server shutdown error", zap.Error(err))
	} else {
		zap.L().Info("Server shutdown completed")
	}
}

// initDatabase инициализирует подключение к PostgreSQL и применяет миграции.
// Параметры:
//   - cfg: конфигурация приложения с DSN строкой подключения
//
// Возвращает:
//   - указатель на репозиторий URL или nil при ошибке
//
// Действия:
//  1. Устанавливает соединение с пулом подключений БД
//  2. Проверяет подключение через ping
//  3. Применяет миграции через стандартную библиотеку database/sql
//  4. Создает репозиторий для работы с URL
func initDatabase(cfg *config.AppConfig) (*repository.Repository, error) {
	pool, connErr := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if connErr != nil {
		return nil, connErr
	}
	if pool.Ping(context.Background()) == nil {
		db := stdlib.OpenDBFromPool(pool)
		if err := postgresql.Migrations(db, "migrations"); err != nil {
			return nil, err
		}
	}

	repo, err := repository.NewRepository(pool)

	if err != nil {
		return nil, err
	}

	return repo, nil
}

func getOrderUseCase(repo *repository.Repository) usecase.OrderUsecase {

	// Инициализация клиента API
	orderClient := orderclient.NewWithDefaults(
		orderclient.WithBaseURL("http://localhost:8081"),
	)

	// Адаптер для клиента
	orderAPI := accrualclient.NewAdapter(orderClient)

	// Инициализация воркера
	orderWorker := worker.NewOrderWorker(worker.DefaultConfig(*orderClient))

	// Инициализация usecase
	orderUseCase := usecase.NewOrderUsecase(*repo, orderWorker, orderAPI)

	return orderUseCase
}

// initServer создает экземпляр HTTP-сервера с использованием фреймворка Chi.
// Параметры:
//   - cfg: конфигурация сервера
//   - handler: обработчик HTTP-запросов
//
// Возвращает:
//   - сконфигурированный экземпляр сервера
//   - завершает приложение при ошибке создания сервера
func initServer(cfg *config.AppConfig, handler *handlers.UserHandler) (*delivery.Server, error) {
	srv, err := delivery.NewServerChi(cfg, handler.ChiMux())
	if err != nil {
		return nil, err
	}
	return srv, nil
}

// loadConfig загружает конфигурацию приложения из YAML-файла.
// Возвращает:
//   - указатель на загруженную конфигурацию
//
// Логика:
//   - Пытается загрузить конфигурацию из файла configs/config.yaml
//   - При ошибке использует конфигурацию по умолчанию
//   - Выполняет валидацию обязательных параметров
//   - Завершает приложение при ошибке валидации
func loadConfig(logger *zap.Logger) (*config.AppConfig, error) {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		cfg = config.NewDefaultConfig()
		logger.Warn("can't load config file, use default config", zap.Error(err))
	}

	if err = cfg.Validate(); err != nil {

		return nil, err
	}

	return cfg, nil
}
