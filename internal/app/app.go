package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/eac0de/gophermart/internal/config"
	"github.com/eac0de/gophermart/internal/database"
	"github.com/eac0de/gophermart/internal/handlers"
	"github.com/eac0de/gophermart/internal/router"
	"github.com/eac0de/gophermart/internal/services"
	"github.com/eac0de/gophermart/pkg/jwt"
	"github.com/eac0de/gophermart/pkg/middlewares"
	"github.com/go-resty/resty/v2"
)

type App struct {
	config   *config.Config
	database *database.Database
	router   *router.AppRouter
}

func NewApp(ctx context.Context, config *config.Config) *App {

	database, err := database.NewDatabase(ctx, config.DatabaseURI)
	if err != nil {
		panic(err)
	}

	tokenService := jwt.NewJWTTokenService(config.SecretKey, 30*time.Minute, 48*time.Hour, database)

	loggerMiddleware := middlewares.GetLoggerMiddleware()
	gzipMiddleware := middlewares.GetGzipMiddleware()
	authMiddleware := middlewares.GetJWTAuthMiddleware(tokenService, database)

	authService := services.NewAuthService(config.SecretKey, database)
	authHandlers := handlers.NewAuthHandlers(tokenService, authService)

	userService := services.NewUserService(database)
	userHandlers := handlers.NewUserHandlers(userService)

	client := resty.New()
	orderService := services.NewOrderService(database, client, config.AccrualSystemAddress)
	go orderService.StartProcessingOrders(ctx)
	orderHandlers := handlers.NewOrderHandlers(orderService)

	balanceService := services.NewBalanceService(database)
	balanceHandlers := handlers.NewBalanceHandlers(balanceService)

	router := router.NewAppRouter(
		orderHandlers,
		authHandlers,
		balanceHandlers,
		userHandlers,
		loggerMiddleware,
		gzipMiddleware,
		authMiddleware,
	)

	return &App{
		config:   config,
		database: database,
		router:   router,
	}
}
func (app *App) Stop() {
	app.database.Close()
	log.Println("Server graceful shutdown")
}

func (app *App) Run() {
	log.Printf("Server http://%s is running. Press Ctrl+C to stop\n", app.config.RunAddress)
	http.ListenAndServe(app.config.RunAddress, app.router)
}
