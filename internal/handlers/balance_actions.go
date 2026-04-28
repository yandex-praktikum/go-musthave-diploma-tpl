package handlers

import (
	"net/http"

	// "github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/common"
	middleware "github.com/KalessinD/gophermart/internal/middleware"
	service "github.com/KalessinD/gophermart/internal/services"
)

type (
	BalanceHandler struct {
		balanceService service.OrderActionsInterface
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
func NewBalancesHandler(balanceService service.OrderActionsInterface) BalanceHandlerInterface {
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
	log := middleware.GetLogger(ctx)

	_ = log

	w.Header().Set("Content-Type", common.AppJSONContentType)
	w.WriteHeader(http.StatusOK)
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
	log := middleware.GetLogger(ctx)

	_ = log

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
	log := middleware.GetLogger(ctx)

	_ = log

	w.Header().Set("Content-Type", common.AppJSONContentType)
	w.WriteHeader(http.StatusOK)
}
