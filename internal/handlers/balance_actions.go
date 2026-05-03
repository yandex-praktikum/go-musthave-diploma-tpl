package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	// "github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/common"
	middleware "github.com/KalessinD/gophermart/internal/middleware"
	"github.com/KalessinD/gophermart/internal/models"
	service "github.com/KalessinD/gophermart/internal/services"
	"go.uber.org/zap"
)

type (
	BalanceHandler struct {
		balanceService service.BalanceServiceInterface
	}

	BalanceHandlerInterface interface {
		GetLoyalityBalance(w http.ResponseWriter, r *http.Request)
		WithdrawBalance(w http.ResponseWriter, r *http.Request)
		ListWithdrawals(w http.ResponseWriter, r *http.Request)
	}
)

/*
Конструктор для хендлеров работающих с балансами
*/
func NewBalancesHandler(balanceService service.BalanceServiceInterface) BalanceHandlerInterface {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

/*
Получение текущего баланса пользователя

Формат запроса:
```
GET /api/user/balance HTTP/1.1
Content-Length: 0
```

Возможные коды ответа:

  - 200 — успешная обработка запроса.
    Формат ответа:
    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    ...

    {
    "current": 500.5,
    "withdrawn": 42
    }

    ```

  - 401 — пользователь не авторизован.

  - 500 — внутренняя ошибка сервера.
*/
func (h *BalanceHandler) GetLoyalityBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx).Sugar()

	balance, err := h.balanceService.GetBalanceInfo(ctx)
	if err != nil {
		status := h.defineResponseStatusByError(err)
		if status == http.StatusInternalServerError {
			log.Errorf("balance service failed to get balance: %s", err.Error())
		} else {
			log.Debugf("balance service failed to get balance: %s", err.Error())
		}

		w.WriteHeader(status)

		return
	}

	body, err := balance.ToJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.AppJSONContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

/*
Запрос на списание средств

Формат запроса:
```
POST /api/user/balance/withdraw HTTP/1.1
Content-Type: application/json

	{
	    "order": "2377225624",
	    "sum": 751
	}

```

Здесь order — номер заказа, а sum — сумма баллов к списанию в счёт оплаты.

Возможные коды ответа:
  - 200 — успешная обработка запроса;
  - 401 — пользователь не авторизован;
  - 402 — на счету недостаточно средств;
  - 422 — неверный номер заказа;
  - 500 — внутренняя ошибка сервера.
*/
func (h *BalanceHandler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx).Sugar()
	contentType := r.Header.Get("Content-Type")

	if contentType != common.AppJSONContentType {
		err := fmt.Errorf("bad request (wrong content-type): %s", contentType)
		log.Error("error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("bad request: %w", err)
		log.Error("error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	withdrawn, err := models.FromJSON[models.Withdrawn](body)
	if err != nil {
		log.Error("error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.balanceService.Withdraw(ctx, withdrawn); err != nil {
		status := h.defineResponseStatusByError(err)
		if status == http.StatusInternalServerError {
			log.Errorf("balance service failed to withdraw: %s", err.Error())
		} else {
			log.Debugf("balance service failed to withdraw: %s", err.Error())
		}

		w.WriteHeader(status)

		return
	}

	w.Header().Set("Content-Type", common.AppJSONContentType)
	w.WriteHeader(http.StatusOK)
}

/*
Получение информации о выводе средств

Формат запроса:
```
GET /api/user/withdrawals HTTP/1.1
Content-Length: 0
```

Возможные коды ответа:
200 — успешная обработка запроса.

	 Формат ответа:
	 ```
	   200 OK HTTP/1.1
	 Content-Type: application/json
	 ...

	 [
	     {
	         "order": "2377225624",
	         "sum": 500,
	         "processed_at": "2020-12-09T16:09:57+03:00"
	     }
	 ]

	 ```
	- 204 — нет ни одного списания.
	- 401 — пользователь не авторизован.
	- 500 — внутренняя ошибка сервера.
*/
func (h *BalanceHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx).Sugar()

	list, err := h.balanceService.ListWithdrawals(ctx)
	if err != nil {
		status := h.defineResponseStatusByError(err)
		if status == http.StatusInternalServerError {
			log.Errorf("balance service failed to list withdrawals: %s", err.Error())
		} else {
			log.Debugf("balance service failed to list withdrawals: %s", err.Error())
		}
		w.WriteHeader(status)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.AppJSONContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (h *BalanceHandler) defineResponseStatusByError(err error) (status int) {
	switch {
	case errors.Is(err, models.ErrOrderWrongFormat):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, models.ErrUserBalanceIsNotEnough):
		status = http.StatusPaymentRequired
	default:
		status = http.StatusInternalServerError
	}

	return
}
