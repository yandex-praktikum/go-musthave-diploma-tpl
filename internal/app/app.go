package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/NailUsmanov/internal/handlers"
	"github.com/NailUsmanov/internal/middleware"
	"github.com/NailUsmanov/internal/storage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type App struct {
	storage storage.Storage
	router  *chi.Mux
	sugar   *zap.SugaredLogger
}

func NewApp(s storage.Storage, sugar *zap.SugaredLogger, accrualHost string) *App {
	r := chi.NewRouter()
	app := &App{
		storage: s,
		router:  r,
		sugar:   sugar,
	}

	sugar.Info("App initialized")
	app.StartWorker(context.Background(), accrualHost)
	app.setupRoutes()
	return app
}

func (a *App) setupRoutes() {
	a.router.Use(middleware.LoggingMiddleWare(a.sugar))
	a.router.Post("/api/user/register", handlers.Register(a.storage, a.sugar))
	a.router.Post("/api/user/login", handlers.Login(a.storage, a.sugar))

	a.router.Route("/api/user", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.GzipMiddleware)
		r.Post("/orders", handlers.PostOrder(a.storage, a.sugar))
		r.Get("/orders", handlers.GetUserOrders(a.storage, a.sugar))
		r.Get("/balance", handlers.UserBalance(a.storage, a.sugar))
		r.Post("/balance/withdraw", handlers.WithDraw(a.storage, a.sugar))
		r.Get("/withdrawals", handlers.AllUserWithDrawals(a.storage, a.sugar))
	})
}
func (a *App) Run(addr string) error {
	return http.ListenAndServe(addr, a.router)
}

// Создаем метод StartWorker для фонового вызова воркера с проверкой статуса заказа из Аккруал хендлера
func (a *App) StartWorker(ctx context.Context, accrualHost string) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				a.sugar.Info("Worker stopped due to context cancellation")
				return
			case <-ticker.C:
				orders, err := a.storage.GetOrdersForAccrualUpdate(ctx)
				a.sugar.Infof(">>> Worker tick: found %d orders", len(orders))
				if err != nil {
					a.sugar.Errorf("Method GetOrdersForAccrualUpdate has err: %v", err)
					continue
				}
				for _, order := range orders {
					// GET запрос в accrual систему: http://{accrualHost}/api/orders/{order.Number}
					url := fmt.Sprintf("%s/api/orders/%s", accrualHost, order.Number)
					resp, err := http.Get(url)
					if err != nil {
						a.sugar.Errorf("HTTP GET failed for %s: %v", url, err)
						continue
					}
					func() {
						defer resp.Body.Close()
						// Обработка ответа
						if resp.StatusCode == http.StatusNoContent {
							return
						}
						// В случае, когда превышено количество запросов, ждем время,
						// которое указно в хедере Retry - After
						if resp.StatusCode == http.StatusTooManyRequests {
							retryAfter := resp.Header.Get("Retry-After")
							if sec, err := strconv.Atoi(retryAfter); err == nil && sec > 0 {
								a.sugar.Warnf("Too many requests, sleeping for %d seconds", sec)
								time.Sleep(time.Duration(sec) * time.Second)
							}
							return
						}
						if resp.StatusCode != http.StatusOK {
							a.sugar.Warnf("Unexpected status from accrual: %d", resp.StatusCode)
							return
						}
						// Создаем структуру аккруал, в которую дальше будем декодировать данные из тела ответа JSON
						var accrualResp struct {
							Order   string   `json:"order"`
							Status  string   `json:"status"`
							Accrual *float64 `json:"accrual,omitempty"`
						}
						if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
							a.sugar.Errorf("Failed to decode accrual response: %v", err)
							return
						}
						// Вызываем метод для обновления данных
						err = a.storage.UpdateOrderStatus(ctx, accrualResp.Order, accrualResp.Status, accrualResp.Accrual)
						if err != nil {
							a.sugar.Errorf("UpdateOrderStatus failed: %v", err)
							return
						}
						a.sugar.Infof("Updated order %s to %s", accrualResp.Order, accrualResp.Status)
					}()
				}
			}
		}
	}()

}
