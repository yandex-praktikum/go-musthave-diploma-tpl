package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
)

func GetBalance() http.Handler {
	balance := func(res http.ResponseWriter, req *http.Request) {
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
		var balanceData storage.BalanceResponce
		balanceData, err = storage.GetUserBalance(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		balanceData.Accrual = balanceData.Accrual / 100
		balanceData.Withdrawn = balanceData.Withdrawn / 100
		result, err := json.Marshal(balanceData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(result)
	}
	return http.HandlerFunc(balance)
}
