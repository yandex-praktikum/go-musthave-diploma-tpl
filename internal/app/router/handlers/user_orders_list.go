package handlers

import (
	resp "github.com/EestiChameleon/GOphermart/internal/app/router/responses"
	"github.com/EestiChameleon/GOphermart/internal/app/service/methods"
	"net/http"
)

// UserOrdersList получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
/*
GET /api/user/orders HTTP/1.1
Content-Length: 0
Возможные коды ответа:
200 — успешная обработка запроса.
Формат ответа:
200 OK HTTP/1.1
Content-Type: application/json
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
204 — нет данных для ответа.
401 — пользователь не авторизован.
500 — внутренняя ошибка сервера.
*/
func UserOrdersList(w http.ResponseWriter, r *http.Request) {
	ordersList, err := methods.GetOrdersListByUserID()
	if err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}

	if len(ordersList) == 0 {
		resp.NoContent(w, http.StatusNoContent)
		return
	}

	resp.JSON(w, http.StatusOK, ordersList)
}
