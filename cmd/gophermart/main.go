package main

import (
	"context"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/storage"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"go.uber.org/zap"
)

// ==================== Константы ====================

const (
	configPath                = "configs/config.yaml"
	migrationsPath            = "migrations"
	defaultHashCost           = 12
	shutdownTimeout           = 10 * time.Second
	pollTimeout               = 5 * time.Second
	defaultSessionTokenExpiry = 1 * time.Hour
	defaultRefreshTokenExpiry = 24 * time.Hour
)

// ==================== Структура приложения ====================

// App представляет основное приложение
type App struct {
	config     *config.AppConfig
	logger     *zap.Logger
	server     *delivery.Server
	repo       *repository.Repository
	userUC     *usecase.UseCase
	orderUC    *usecase.OrderUseCase
	worker     worker.OrderWorker
	cancelFunc context.CancelFunc
	shutdownCh chan struct{}
}

// NewApp создает новое приложение
func NewApp() (*App, error) {
	// Инициализация логгера
	logger, err := initLogger()
	if err != nil {
		return nil, err
	}

	// Загрузка конфигурации
	cfg, err := loadConfig(logger)
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		return nil, err
	}

	return &App{
		config:     cfg,
		logger:     logger,
		shutdownCh: make(chan struct{}),
	}, nil
}

// ==================== Инициализация компонентов ====================

// initLogger инициализирует zap логгер
func initLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, err
	}

	// Заменяем глобальный логгер
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// loadConfig загружает конфигурацию приложения
func loadConfig(logger *zap.Logger) (*config.AppConfig, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Warn("Can't load config file, using default config",
			zap.String("config_path", configPath),
			zap.Error(err),
		)
		cfg = config.NewDefaultConfig()
	}

	// Валидация конфигурации
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration", zap.Error(err))
		return nil, err
	}

	logger.Info("Configuration loaded successfully",
		zap.String("run_address", cfg.RunAddress),
		zap.String("database_uri", maskDatabaseURI(cfg.DatabaseURI)),
		zap.String("accrual_address", cfg.AccrualSystemAddress),
	)

	return cfg, nil
}

// maskDatabaseURI маскирует чувствительные данные в URI базы данных
func maskDatabaseURI(uri string) string {
	// Простая маскировка пароля в DSN
	// В реальном приложении используйте более сложную логику
	return uri
}

// Init инициализирует все компоненты приложения
func (a *App) Init() error {
	a.logger.Info("Initializing application components...")

	// Инициализация базы данных
	repo, err := a.initDatabase()
	if err != nil {
		a.logger.Error("Failed to initialize database", zap.Error(err))
		return err
	}
	a.repo = repo

	// Инициализация use cases
	if err := a.initUseCases(); err != nil {
		a.logger.Error("Failed to initialize use cases", zap.Error(err))
		return err
	}

	// Инициализация сервера
	if err := a.initServer(); err != nil {
		a.logger.Error("Failed to initialize server", zap.Error(err))
		return err
	}

	a.logger.Info("Application initialized successfully")
	return nil
}

// initDatabase инициализирует подключение к базе данных
func (a *App) initDatabase() (*repository.Repository, error) {
	a.logger.Info("Initializing database connection...",
		zap.String("database_uri", maskDatabaseURI(a.config.DatabaseURI)),
	)

	// Создание пула подключений
	pool, err := pgxpool.New(context.Background(), a.config.DatabaseURI)
	if err != nil {
		a.logger.Error("Failed to create database pool", zap.Error(err))
		return nil, err
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		a.logger.Error("Database ping failed", zap.Error(err))
		return nil, err
	}

	a.logger.Info("Database connection established")

	// Применение миграций
	a.logger.Info("Applying database migrations...",
		zap.String("migrations_path", migrationsPath),
	)

	db := stdlib.OpenDBFromPool(pool)
	if err := postgresql.Migrations(db, migrationsPath); err != nil {
		a.logger.Error("Failed to apply migrations", zap.Error(err))
		return nil, err
	}

	a.logger.Info("Database migrations applied successfully")

	// Создание репозитория
	repo, err := repository.NewRepository(pool)
	if err != nil {
		a.logger.Error("Failed to create repository", zap.Error(err))
		return nil, err
	}

	a.logger.Info("Repository initialized")
	return repo, nil
}

// initUseCases инициализирует use cases
func (a *App) initUseCases() error {
	a.logger.Info("Initializing use cases...")

	// Инициализация клиента заказов
	orderClient := orderclient.NewWithDefaults(
		orderclient.WithBaseURL(a.config.AccrualSystemAddress),
		orderclient.WithTimeout(pollTimeout),
	)

	a.logger.Info("Order client initialized",
		zap.String("base_url", a.config.AccrualSystemAddress),
	)

	// Инициализация воркера
	workerConfig := worker.DefaultConfig(orderClient, a.repo, a.logger)
	a.worker = worker.NewOrderWorker(workerConfig)

	a.logger.Info("Order worker initialized",
		zap.Int("workers", workerConfig.Workers),
		zap.Duration("interval", workerConfig.Interval),
	)

	// Инициализация use case для пользователей
	a.userUC = usecase.NewUseCase(
		a.repo,
		a.config.SecretKey,
		defaultHashCost,
		defaultSessionTokenExpiry,
		defaultRefreshTokenExpiry,
	)

	// Инициализация use case для заказов
	a.orderUC = usecase.NewOrderUseCase(a.repo, a.worker)

	a.logger.Info("Use cases initialized successfully")
	return nil
}

// initServer инициализирует HTTP-сервер
func (a *App) initServer() error {
	a.logger.Info("Initializing HTTP server...",
		zap.String("address", a.config.RunAddress),
	)

	// Создаем обработчик (нужно получить роутер от обработчика)
	// В реальной реализации нужно получить Chi роутер
	handler := handlers.NewUserHandler(
		a.userUC,
		a.orderUC,
		a.config.RunAddress,
		a.config.AccrualSystemAddress,
		storage.NewSessionStore(a.logger),
		a.logger,
	)

	// Создаем сервер с Chi роутером
	server, err := delivery.NewServerChi(a.config, handler.ChiMux())
	if err != nil {
		a.logger.Error("Failed to create server", zap.Error(err))
		return err
	}

	a.server = server
	a.logger.Info("HTTP server initialized")
	return nil
}

// ==================== Управление жизненным циклом ====================

// Start запускает приложение
func (a *App) Start() error {
	a.logger.Info("Starting application...")

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel

	// Запускаем обработку заказов
	a.logger.Info("Starting order processing...")
	if err := a.orderUC.StartOrderProcessing(ctx); err != nil {
		a.logger.Error("Failed to start order processing", zap.Error(err))
		return err
	}

	// Запускаем HTTP-сервер
	a.logger.Info("Starting HTTP server...")
	serverErrChan := a.server.Start()

	// Обрабатываем сигналы и ошибки сервера
	go a.handleSignals(ctx)
	go a.handleServerErrors(ctx, serverErrChan)

	a.logger.Info("Application started successfully",
		zap.String("address", a.config.RunAddress),
	)

	return nil
}

// handleSignals обрабатывает сигналы завершения
func (a *App) handleSignals(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Ожидаем сигналы в цикле
	for {
		select {
		case <-ctx.Done():
			a.logger.Debug("Signal handler stopped (context canceled)")
			return
		case sig := <-sigChan:
			a.logger.Info("Received shutdown signal",
				zap.String("signal", sig.String()),
			)

			// Вызываем shutdown и выходим из цикла
			a.Shutdown()
			return
		}
	}
}

// handleServerErrors обрабатывает ошибки сервера
func (a *App) handleServerErrors(ctx context.Context, errChan <-chan error) {
	for {
		select {
		case <-ctx.Done():
			a.logger.Debug("Server error handler stopped (context canceled)")
			return
		case err := <-errChan:
			if err != nil {
				a.logger.Error("Server error", zap.Error(err))
				a.Shutdown()
				return
			}
		}
	}
}

// Shutdown выполняет graceful shutdown приложения
func (a *App) Shutdown() {
	a.logger.Info("Initiating graceful shutdown...")

	// Отменяем контекст для остановки всех горутин
	if a.cancelFunc != nil {
		a.cancelFunc()
	}

	// Создаем контекст с таймаутом для всего shutdown процесса
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем обработку заказов
	if a.orderUC != nil {
		a.logger.Info("Stopping order processing...")
		if err := a.orderUC.StopOrderProcessing(shutdownCtx); err != nil {
			a.logger.Error("Failed to stop order processing", zap.Error(err))
		} else {
			a.logger.Info("Order processing stopped")
		}
	}

	// Останавливаем HTTP-сервер
	if a.server != nil {
		a.logger.Info("Shutting down HTTP server...")

		if err := a.server.Shutdown(15 * time.Second); err != nil {
			a.logger.Error("Failed to shutdown server", zap.Error(err))
		} else {
			a.logger.Info("HTTP server shutdown completed")
		}
	}

	// Останавливаем воркер (если он имеет метод Stop)
	if worker, ok := a.worker.(interface{ Stop() }); ok {
		a.logger.Info("Stopping worker...")
		worker.Stop()
		a.logger.Info("Worker stopped")
	}

	// Закрываем соединение с базой данных
	if a.repo != nil {
		a.logger.Info("Closing database connections...")
		a.repo.Close()
		a.logger.Info("Database connections closed")
	}

	// Синхронизируем логгер
	a.logger.Info("Syncing logger...")
	_ = a.logger.Sync()

	a.logger.Info("Application shutdown completed")

	// Сигнализируем о завершении shutdown
	close(a.shutdownCh)
}

// WaitForShutdown ожидает завершения shutdown
func (a *App) WaitForShutdown() {
	<-a.shutdownCh
}

// ==================== Точка входа ====================

func main() {
	// Создание приложения
	app, err := NewApp()
	if err != nil {
		// Используем стандартный логгер, если zap не инициализирован
		if app != nil && app.logger != nil {
			app.logger.Fatal("Failed to create application", zap.Error(err))
		} else {
			panic(err)
		}
	}

	// Инициализация компонентов
	if err := app.Init(); err != nil {
		app.logger.Fatal("Failed to initialize application", zap.Error(err))
	}

	// Запуск приложения
	if err := app.Start(); err != nil {
		app.logger.Fatal("Failed to start application", zap.Error(err))
	}

	// Ждем завершения shutdown
	app.WaitForShutdown()

	app.logger.Info("Application terminated successfully")
}
