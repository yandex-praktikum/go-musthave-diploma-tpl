package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"musthave/internal/model"
	"net/http"
	"time"
)

const (
	order = "/api/orders/"
)

type Client struct {
	baseURL      string
	client       *http.Client
	attemptCount int
}

func NewClient(path string, timeOut time.Duration) *Client {
	return &Client{
		baseURL: path,
		client: &http.Client{
			Timeout: timeOut,
		},
	}
}

func (c *Client) GetAccrual(ctx context.Context, lg *slog.Logger, orderID int) (*model.AccrualRes, error) {
	path := c.baseURL + order + fmt.Sprintf("%v", orderID) //?
	var count int
	for {
		if count >= c.attemptCount {
			return nil, fmt.Errorf("превышено макс.количество попыток")
		}
		lg.Info("GetAccrual.start - старт новой иттерации на запрос статуса по заказу: " + fmt.Sprintf("%v", orderID))
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			lg.Error("GetAccrual.err - ошибка создания запроса: " + err.Error() + ". Пробуем еще раз")
			//return nil, fmt.Errorf("ошибка создания запроса: %w", err)
			count++
			continue
		}
		resp, err := c.client.Do(req)
		if err != nil {
			lg.Error("GetAccrual.err - ошибка HTTP-запроса:" + err.Error() + ". Пробуем еще раз")
			count++
			continue
			//return nil, fmt.Errorf("ошибка HTTP-запроса: %w", err)
		}
		defer resp.Body.Close()

		var result model.AccrualRes

		switch resp.StatusCode {
		case http.StatusOK:
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				lg.Error("GetAccrual.err - ошибка парсинга JSON: " + err.Error() + ". Пробуем еще раз")
				count++
				continue
			}
			return &result, nil
		case http.StatusNoContent:
			lg.Error("GetAccrual.err - заказ не зарегистрирован в системе расчёта. Пробуем еще раз")
			count++
			continue
			//return nil, fmt.Errorf("заказ не зарегистрирован в системе расчёта")
		case http.StatusInternalServerError:
			lg.Error("GetAccrual.err - внутренняя ошибка сервера расчета. Пробуем еще раз")
			count++
			continue
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("превышено количество запросов к сервису")
		default:
			return nil, fmt.Errorf("неизвестная статус-код")
		}
	}
}
