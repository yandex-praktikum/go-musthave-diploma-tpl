package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	// "github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/common"
	middleware "github.com/KalessinD/gophermart/internal/middleware"
	model "github.com/KalessinD/gophermart/internal/models"
	service "github.com/KalessinD/gophermart/internal/services"
)

type (
	OrdersdHandler struct {
		orderService service.OrderActionsInterface
	}

	RestrictedHandlerInterface interface {
		AddOrder(w http.ResponseWriter, r *http.Request)
		ListOrders(w http.ResponseWriter, r *http.Request)
	}
)

/*
Конструктор для хендлеров работающих с заказами
*/
func NewOrdersHandler(orderService service.OrderActionsInterface) RestrictedHandlerInterface {
	return &OrdersdHandler{
		orderService: orderService,
	}
}

/*
Загрузка номера заказа

Формат запроса:
```
POST /api/user/orders HTTP/1.1
Content-Type: text/plain
...

12345678903
````

Возможные коды ответа:
  - 200 — номер заказа уже был загружен этим пользователем;
  - 202 — новый номер заказа принят в обработку;
  - 400 — неверный формат запроса;
  - 401 — пользователь не аутентифицирован;
  - 409 — номер заказа уже был загружен другим пользователем;
  - 422 — неверный формат номера заказа;
  - 500 — внутренняя ошибка сервера.

Номер заказа проверяется на корректность с помощью алгоритма Луна.
*/
func (h *OrdersdHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx).Sugar()

	contentType := r.Header.Get("Content-Type")
	if contentType != common.TextPlainContentType {
		err := fmt.Errorf("bad request: %s", "wrong content-type")
		log.Debugf("bad request: content-type is %s: %s", contentType, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("bad request: %w", err)
		log.Debugf("bad request: content-type is %s: %s", contentType, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.orderService.Store(ctx, string(body)); err != nil {
		status := h.defineResponseStatusByError(err)
		if status == http.StatusInternalServerError {
			log.Errorf("order service failed: %s", err.Error())
		} else {
			log.Debugf("order service failed: %s", err.Error())
		}

		w.WriteHeader(status)

		return
	}

	w.Header().Set("Content-Type", common.TextPlainContentType)
	w.WriteHeader(http.StatusOK)
}

/*
Получение списка загруженных номеров заказов

Формат запроса:
````
GET /api/user/orders HTTP/1.1
Content-Length: 0
````

Возможные коды ответа:
  - 200 — успешная обработка запроса.
    Формат ответа:

```

		 200 OK HTTP/1.1
		 Content-Type: application/json
		 ...

		 [
		     {
		         "number": "9278923470",
		         "status": "PROCESSED",
		         "accrual": 500,
		         "uploaded_at": "2020-12-10T15:15:45+03:00"
		     },
		     {
		         "number": "12345678903",
		         "status": "PROCESSING",
		         "uploaded_at": "2020-12-10T15:12:01+03:00"
		     },
		     {
		         "number": "346436439",
		         "status": "INVALID",
		         "uploaded_at": "2020-12-09T16:09:53+03:00"
		     }
		 ]
		```

	 - 204 — нет данных для ответа.
	 - 401 — пользователь не авторизован.
	 - 500 — внутренняя ошибка сервера.
*/
func (h *OrdersdHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := middleware.GetLogger(ctx)

	_ = log

	w.Header().Set("Content-Type", common.AppJSONContentType)
	w.WriteHeader(http.StatusOK)
}

func (h *OrdersdHandler) defineResponseStatusByError(err error) (status int) {
	switch {
	case errors.Is(err, model.ErrOrderNotFound):
		status = http.StatusNoContent
	default:
		status = http.StatusInternalServerError
	}

	return
}
