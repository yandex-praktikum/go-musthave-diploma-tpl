package handlers

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"time"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
	"github.com/Azcarot/GopherMarketProject/internal/utils"
)

type OrderRequest struct {
	OrderNumber int
}

func Order() http.Handler {
	order := func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		token := req.Header.Get("Authorization")
		claims, ok := storage.VerifyToken(token)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
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
		// читаем тело запроса
		_, err = buf.ReadFrom(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		data := buf.Bytes()
		orderNumber := binary.BigEndian.Uint64(data)
		ok = utils.IsOrderNumberValid(orderNumber)
		if !ok {
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		var order storage.OrderData
		order.OrderNumber = orderNumber
		order.User = userData.Login
		ok, anotherUser := storage.CheckIfOrderExists(storage.DB, order)
		if anotherUser {
			res.WriteHeader(http.StatusConflict)
			return
		}
		if !ok {
			res.WriteHeader(http.StatusOK)
			return
		}
		order.Date = time.Now()
		err = storage.CreateNewOrder(storage.DB, order)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusAccepted)

	}
	return http.HandlerFunc(order)
}
