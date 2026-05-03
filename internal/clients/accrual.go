//go:generate mockgen -source=accrual.go -destination=mocks/mock_accrual_client.go -package=mocks
package clients

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

const (
	// DefaultDelay int = 60 // seconds

	RetryingAttempts  = 3
	RetryingDelay     = 100 * time.Microsecond
	RetryingDelayStep = 200 * time.Microsecond

	BaseAPIOrderURL = "/api/orders/"

	DefaultDelay = 10
)

type (
	AccrualClient struct {
		BaseClient *http.Client
		BaseURL    string
		Log        *zap.Logger
	}

	AccrualClienttInterface interface {
		GetOrderAccrual(ctx context.Context, orderID string) (*models.AccrualResponse, error)
	}
)

// Конструктор http-клиента для запросов в систему Accrual
func NewAccrualClient(accrualAddress string, timeout time.Duration, log *zap.Logger) AccrualClienttInterface {
	return &AccrualClient{
		BaseClient: &http.Client{Timeout: timeout},
		BaseURL:    accrualAddress,
		Log:        log,
	}
}

// Выполняет запрос с попытками реконнекта
func (c *AccrualClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	attempts := RetryingAttempts
	delay := RetryingDelay

	for attempts > 0 {
		// nolint:gosec
		response, err := c.BaseClient.Do(req)

		if err == nil {
			return response, nil
		}

		if attempts == 1 {
			return nil, err
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		attempts--
		delay += RetryingDelayStep
	}

	return nil, errors.New("no response")
}

/*

Для взаимодействия с системой доступен один хендлер:
GET /api/orders/{number} — получение информации о расчёте начислений баллов лояльности.

Формат запроса:

````
GET /api/orders/{number} HTTP/1.1
Content-Length: 0
```
`
Возможные коды ответа:
- 200 — успешная обработка запроса.
  Формат ответа:
```
  200 OK HTTP/1.1
  Content-Type: application/json
	...

  {
      "order": "<number>",
      "status": "PROCESSED",
      "accrual": 500
  }
```
  Поля объекта ответа:
	- order — номер заказа;
	- status — статус расчёта начисления:
		- REGISTERED — заказ зарегистрирован, но вознаграждение не рассчитано;
		- INVALID — заказ не принят к расчёту, и вознаграждение не будет начислено;
		- PROCESSING — расчёт начисления в процессе;
		- PROCESSED — расчёт начисления окончен;
	- accrual — рассчитанные баллы к начислению, при отсутствии начисления — поле отсутствует в ответе.
- 204 — заказ не зарегистрирован в системе расчёта.
- 429 — превышено количество запросов к сервису.
	  Формат ответа:
```
  429 Too Many Requests HTTP/1.1
  Content-Type: text/plain
  Retry-After: 60
`
  No more than N requests per minute allowed
```
- 500 — внутренняя ошибка сервера.

Заказ может быть взят в расчёт в любой момент после его совершения.
Время выполнения расчёта системой не регламентировано.
Статусы INVALID и PROCESSED являются окончательными.
Общее количество запросов информации о начислении не ограничено.

*/

func (c *AccrualClient) GetOrderAccrual(ctx context.Context, orderID string) (*models.AccrualResponse, error) {
	url := c.BaseURL + "/api/orders/" + orderID
	req, err := c.prepareRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	c.Log.Info("Request to accrual service was sent",
		zap.String("request_id", req.Header.Get("X-Request-Id")),
		zap.String("url", url),
		zap.Int("response status", resp.StatusCode),
	)

	switch resp.StatusCode {
	case http.StatusOK:
		return c.parseAccrualResponse(resp)

	case http.StatusNoContent:
		return nil, ErrOrderNotFound

	case http.StatusTooManyRequests:
		if resp.StatusCode == http.StatusTooManyRequests {
			var delay int
			var err error
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				delay, err = strconv.Atoi(retryAfter)
			}
			if err != nil || delay == 0 {
				delay = DefaultDelay
			}
			return nil, NewErrServiceIsBusy(time.Duration(delay)*time.Second, errors.New(resp.Status))
		}

	default:
		// 500, 400 и прочие
	}

	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func (c *AccrualClient) prepareRequest(ctx context.Context, url string) (*http.Request, error) {
	requestID := middleware.GetReqID(ctx)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Accept", common.AppJSONContentType)

	return req, nil
}

func (c *AccrualClient) parseAccrualResponse(resp *http.Response) (*models.AccrualResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read a response body: %w", err)
	}

	accrualResp, err := models.FromJSON[models.AccrualResponse](body)
	if err != nil {
		return nil, fmt.Errorf("can't parse a response body: %w", err)
	}

	return accrualResp, nil
}
