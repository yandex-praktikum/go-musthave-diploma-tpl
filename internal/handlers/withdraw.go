package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
	"github.com/Azcarot/GopherMarketProject/internal/utils"
)

func Withdraw() http.Handler {
	withdraw := func(res http.ResponseWriter, req *http.Request) {
		// var buf bytes.Buffer
		token := req.Header.Get("Authorization")
		claims, ok := storage.VerifyToken(token)
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		var userData storage.UserData
		userData.Login = claims["sub"].(string)
		ok, err := storage.CheckUserExists(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		var buf bytes.Buffer
		withdrawalData := storage.WithdrawRequest{}
		// читаем тело запроса
		_, err = buf.ReadFrom(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		data := buf.Bytes()

		if err = json.Unmarshal(data, &withdrawalData); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		orderNumber, err := strconv.ParseUint(withdrawalData.OrderNumber, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ok = utils.IsOrderNumberValid(orderNumber)
		if !ok {
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		var balanceData storage.BalanceResponce
		balanceData, err = storage.GetUserBalance(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		intAccBalanceData := int(balanceData.Accrual * 100)
		intWithdData := int(balanceData.Withdrawn * 100)
		if int(withdrawalData.Amount*100) > intAccBalanceData {
			res.WriteHeader(http.StatusPaymentRequired)
			return
		}
		userData.AccrualPoints = intAccBalanceData
		userData.Withdrawal = intWithdData
		err = storage.WitdrawFromUser(storage.DB, userData, withdrawalData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(withdraw)
}
