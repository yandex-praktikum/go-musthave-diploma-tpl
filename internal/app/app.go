// Package app содержит основной контейнер приложения
package app

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/anon-d/gophermarket/internal/config"
	"github.com/anon-d/gophermarket/internal/http/handler"
	"github.com/anon-d/gophermarket/internal/http/middleware"
	"github.com/anon-d/gophermarket/internal/http/service"
	"github.com/anon-d/gophermarket/internal/logger"
	"github.com/anon-d/gophermarket/internal/repository/postgres"
	"github.com/anon-d/gophermarket/internal/worker"
)

const (
	// DefaultNamespace это UUID неймспейс для генерации user_id
	// зашил так как его нет в обязательных параметрах запуска,
	// а без него не сгенерируется UUID для ID в бд
	DefaultNamespace = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
)

type Application struct {
	server        *http.Server
	router        *gin.Engine
	handler       *handler.GopherHandler
	logger        *zap.Logger
	repo          *postgres.PostgresDB
	accrualWorker *worker.AccrualWorker
	cfg           *config.Config
	workerCancel  context.CancelFunc
}

func New() *Application {
	cfg, err := config.New()
	if err != nil {
		log.Printf("Ошибка инициализации конфига: %v", err)
		return nil
	}

	zapLogger, err := logger.New()
	if err != nil {
		log.Printf("Ошибка инициализации логгера: %v", err)
		return nil
	}

	// Инициализация БД
	repo, err := postgres.NewPostgres(cfg.DSN, zapLogger)
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		return nil
	}

	// Инициализация сервиса
	jwtSecret := cfg.PrivateKeyJWT
	if jwtSecret == "" {
		jwtSecret = "default-secret-key" // зашил по той же причине. Необходим для генерации токена
	}
	svc := service.NewGopherService(DefaultNamespace, repo, zapLogger, jwtSecret)

	// Инициализация хендлера
	h := handler.NewGopherHandler(svc)

	// Инициализация accrual worker
	var accrualWorker *worker.AccrualWorker
	if cfg.AccrualSystemAddress != "" {
		accrualWorker = worker.NewAccrualWorker(repo, cfg.AccrualSystemAddress, zapLogger)
	}

	// Инициализация Gin
	if cfg.ENV == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.Default()

	server := &http.Server{
		Addr:    cfg.SelfAddress,
		Handler: router,
	}

	return &Application{
		logger:        zapLogger,
		router:        router,
		server:        server,
		handler:       h,
		repo:          repo,
		accrualWorker: accrualWorker,
		cfg:           cfg,
	}
}

// Run запускает приложение
func (a *Application) Run() error {
	a.setupRouter()

	// Запускаем accrual worker
	if a.accrualWorker != nil {
		ctx, cancel := context.WithCancel(context.Background())
		a.workerCancel = cancel
		go a.accrualWorker.Start(ctx)
		a.logger.Info("Запущен accrual worker")
	}

	a.logger.Info("Сервер запущен", zap.String("address", a.server.Addr))
	return a.server.ListenAndServe()
}

// Shutdown останавливает приложение
// Останавливаем воркер -> перестаем принимать запросы -> закрываем таски с БД -> OK
func (a *Application) Shutdown(ctx context.Context) error {
	// Останавливаем worker
	if a.workerCancel != nil {
		a.workerCancel()
	}

	// Останавливаем HTTP сервер
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	// Закрываем БД
	if a.repo != nil {
		if err := a.repo.Close(); err != nil {
			a.logger.Error("Ошибка закрытия БД", zap.Error(err))
		}
	}

	a.logger.Info("Приложение остановлено")
	return nil
}

func (a *Application) setupRouter() {
	// CORS middleware
	a.router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
	}))

	// Security headers
	a.router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	})

	// JWT secret для middleware
	jwtSecret := a.cfg.PrivateKeyJWT
	if jwtSecret == "" {
		jwtSecret = "default-secret-key"
	}

	api := a.router.Group("/api/user")
	{
		// Публичные эндпоинты (без авторизации)
		api.POST("/register", a.handler.RegisterUser)
		api.POST("/login", a.handler.LoginUser)

		// Защищённые эндпоинты (с авторизацией)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.POST("/orders", a.handler.CreateOrder)
			protected.GET("/orders", a.handler.GetOrders)
			protected.GET("/balance", a.handler.GetBalance)
			protected.POST("/balance/withdraw", a.handler.WithdrawBalance)
			protected.GET("/withdrawals", a.handler.GetWithdrawals)
		}
	}

	a.router.HandleMethodNotAllowed = true
}
