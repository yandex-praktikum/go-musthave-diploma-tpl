package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
)

func GetBalance(res http.ResponseWriter, req *http.Request) {
	var userData storage.UserData
	var ctx context.Context
	ctx = context.Background()
	data, ok := req.Context().Value("UserLogin").(string)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	userData.Login = data
	ctxUserKey := storage.UserLoginCtxKey
	ctx = context.WithValue(ctx, ctxUserKey, userData.Login)
	var balanceData storage.BalanceResponce
	balanceData, err := storage.GetUserBalance(storage.DB, userData, ctx)
	if err != nil {
		fmt.Println(err)
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
