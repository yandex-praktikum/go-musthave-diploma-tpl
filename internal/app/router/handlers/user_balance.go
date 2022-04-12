package handlers

import (
	resp "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/router/responses"
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/service/methods"
	"net/http"
)

// UserBalance предоставляет возможность получения текущего баланса счёта баллов лояльности пользователя
/*
GET /api/user/balance HTTP/1.1
Content-Length: 0
Возможные коды ответа:
200 — успешная обработка запроса.
Формат ответа:
200 OK HTTP/1.1
Content-Type: application/json
{
	"current": 500.5,
	"withdrawn": 42
}
401 — пользователь не авторизован.
500 — внутренняя ошибка сервера.
*/
func UserBalance(w http.ResponseWriter, r *http.Request) {
	b := methods.NewBalanceRecord()

	if res, err := b.GetBalanceAndWithdrawnByUserID(); err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	} else {
		resp.JSON(w, http.StatusOK, res)
	}
}
