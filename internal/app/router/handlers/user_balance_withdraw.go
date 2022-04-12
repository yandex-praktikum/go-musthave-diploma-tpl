package handlers

import (
	"encoding/json"
	resp "github.com/EestiChameleon/GOphermart/internal/app/router/responses"
	"github.com/EestiChameleon/GOphermart/internal/app/service"
	"github.com/EestiChameleon/GOphermart/internal/app/service/methods"
	"io"
	"net/http"
)

// UserBalanceWithdraw запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
/*
Формат запроса:
POST /api/user/balance/withdraw HTTP/1.1
Content-Type: application/json
{
    "order": "2377225624",
    "sum": 751
}
Здесь order — номер заказа, а sum — сумма баллов к списанию в счёт оплаты.
Возможные коды ответа:
200 — успешная обработка запроса;
401 — пользователь не авторизован;
402 — на счету недостаточно средств;
422 — неверный номер заказа;
500 — внутренняя ошибка сервера.
*/
func UserBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	var b resp.WithdrawData
	data, err := io.ReadAll(r.Body)
	if err != nil {
		resp.NoContent(w, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, &b)
	if err != nil {
		resp.NoContent(w, http.StatusBadRequest)
		return
	}

	if b.Order == "" || !service.LuhnCheck(b.Order) {
		resp.NoContent(w, http.StatusUnprocessableEntity)
		return
	}

	// get current balance and whole withdrawn
	blnc := methods.NewBalanceRecord()
	res, err := blnc.GetBalanceAndWithdrawnByUserID()
	if err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}

	if !res.Current.GreaterThanOrEqual(b.Sum) {
		resp.NoContent(w, http.StatusPaymentRequired)
		return
	}

	// withdrawn record save
	blnc.Outcome = b.Sum
	blnc.OrderNumber = b.Order
	if err = blnc.Add(); err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}

	resp.NoContent(w, http.StatusOK)
}
