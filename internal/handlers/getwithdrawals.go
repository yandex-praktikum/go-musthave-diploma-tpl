package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
)

func GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	var userData storage.UserData
	data, ok := req.Context().Value("UserLogin").(string)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	userData.Login = data
	withdrawals, err := storage.GetWithdrawals(storage.DB, userData)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(withdrawals)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(result)
}
