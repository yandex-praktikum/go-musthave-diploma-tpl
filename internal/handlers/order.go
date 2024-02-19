package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
	"github.com/Azcarot/GopherMarketProject/internal/utils"
)

type OrderRequest struct {
	OrderNumber uint64
	Status      string
	Accural     int
}

func Order(flag utils.Flags) http.Handler {
	order := func(res http.ResponseWriter, req *http.Request) {
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
		// читаем тело запроса
		data, err := io.ReadAll(req.Body)
		asString := string(data)
		// _, err = buf.ReadFrom(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// data := buf.Bytes()
		orderNumber, err := strconv.ParseUint(asString, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
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
		orderAccData, err := GetOrderAccuralAndState(flag, order.OrderNumber)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		order.Accural = orderAccData.Accural
		order.State = orderAccData.Status
		order.Date = time.Now().Format(time.RFC3339)
		err = storage.CreateNewOrder(storage.DB, order)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusAccepted)

	}
	return http.HandlerFunc(order)
}
