package serve

import (
	"context"
	"errors"
	"fmt"
	"gmart/internal/adapters/metrics"
	"gmart/internal/adapters/pgc"
	"gmart/internal/config"
	"gmart/internal/model/auth"
	"gmart/internal/service/loyalty"
	"gmart/internal/service/orders"
	"gmart/internal/service/user"
	"net/http"
	"time"

	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
)

// Input входные параметры для Serve
type Input struct {
	Cfg        config.Config
	Pg         pgc.PgInstance
	MetricsReg *prometheus.Registry
}

// Запуск сервера. Передаем конфиг и роутер
func Serve(ctx context.Context, arg *Input) error {
	slog.Info("Starting server",
		slog.String(
			"ListenAddr",
			arg.Cfg.ListenAddr(),
		),
		slog.String(
			"AccrualSystemAddr",
			arg.Cfg.AccrualSystemAddr(),
		),
		slog.String(
			"PgInstance",
			arg.Pg.String(),
		),
	)

	mux := http.NewServeMux()
	humaAPI := InitHuma(mux)

	user, err := user.NewUser(arg.Cfg, arg.Pg, metrics.NewForAuth(arg.MetricsReg))
	if err != nil {
		return fmt.Errorf("create user fail: %w", err)
	}

	orders, err := orders.NewOrders(arg.Cfg, arg.Pg, metrics.NewForOrders(arg.MetricsReg))
	if err != nil {
		return fmt.Errorf("create order fail: %w", err)
	}

	loyalty, err := loyalty.NewLoyalty(arg.Cfg, arg.Pg, metrics.NewForLoyalty(arg.MetricsReg))
	if err != nil {
		return fmt.Errorf("create loyalty fail: %w", err)
	}

	tokenVerifier, err := auth.NewTokenVerifier(jwt.SigningMethodHS256, arg.Cfg.JWTSecretKey())
	if err != nil {
		return fmt.Errorf("failed to create token verifier: %w", err)
	}

	// Регистрация роутингов
	user.RegistryRoutes(humaAPI)
	orders.RegistryRoutes(humaAPI, tokenVerifier)
	loyalty.RegistryRoutes(humaAPI, tokenVerifier)

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:              arg.Cfg.ListenAddr(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // Защита от Slowloris атак
		IdleTimeout:       30 * time.Second,
	}

	// Канал для ошибок сервера
	serverErrors := make(chan error, 1)

	// Запускаем сервер в горутине
	go func() {
		slog.Info("Server is ready to handle requests")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("listen and serve: %w", err)
			close(serverErrors)
		}
	}()

	// Блокируемся и ждем либо ошибку, либо сигнал отмены контекста
	select {
	case err := <-serverErrors:
		return err

	case <-ctx.Done():
		slog.Info("Shutting down server...")

		// Даем серверу 5 секунд на завершение текущих запросов
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("forced shutdown: %w", err)
		}
		slog.Info("Server stopped gracefully")
	}

	return nil
}

// InitHuma возвращает инициализированный HUMA
func InitHuma(mux *http.ServeMux) huma.API {
	slog.Info("Init huma router")

	// API with huma
	humaConfig := huma.DefaultConfig("GopherMart API", "1.0.0")

	// humaConfig.DocsRenderer = huma.DocsRendererSwaggerUI

	humaConfig.Formats["text/plain"] = huma.DefaultFormats["text/plain"]

	humaConfig.Components = &huma.Components{
		SecuritySchemes: map[string]*huma.SecurityScheme{
			"bearer": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "Введите JWT токен для доступа к защищенным ресурсам.",
			},
		},
	}

	api := humago.New(mux, humaConfig)

	return api
}
