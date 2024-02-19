package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
	"github.com/Azcarot/GopherMarketProject/internal/utils"
	"github.com/jackc/pgx/v5"
)

func GetOrders() http.Handler {
	getorder := func(res http.ResponseWriter, req *http.Request) {
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
		orders, err := storage.GetCustomerOrders(storage.DB, userData.Login)
		if err == pgx.ErrNoRows {
			res.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		result, err := json.Marshal(orders)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write(result)

	}
	return http.HandlerFunc(getorder)
}
func GetOrderData(flag utils.Flags, order uint64) (*http.Response, error) {
	pth := "http://" + flag.FlagAccuralAddr + "/api/orders/" + strconv.Itoa(int(order))
	var b []byte
	resp, err := http.NewRequest("GET", pth, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(resp)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return res, nil
}

func GetOrderAccuralAndState(flag utils.Flags, order uint64) (OrderRequest, error) {
	result := OrderRequest{}
	response, err := GetOrderData(flag, order)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	var buf bytes.Buffer
	// читаем тело запроса
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return result, err
	}
	data := buf.Bytes()

	if err = json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}
