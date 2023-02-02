package events

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"GopherMart/cmd/errorsGM"
)

type requestAccrualFloat struct {
	Number     string  `json:"number"`
	Status     string  `json:"Status"`
	Point      float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type requestAccrual struct {
	Number     string `json:"number"`
	Status     string `json:"Status"`
	Point      uint   `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}

func accrualOrderStatus(storageAccrual string, order string) (bodyRequest requestAccrual, err error) {
	var errGet *errorsGM.ErrorGopherMart
	accrual, sec, errGet := accrualGet(storageAccrual, order)
	for ; sec != 0; accrual, sec, err = accrualGet(storageAccrual, order) {
		time.Sleep(time.Duration(sec) * time.Second)
	}
	if errGet != nil {
		return requestAccrual{}, errGet
	}

	return accrual, nil
}

func accrualGet(storage string, order string) (bodyRequest requestAccrual, duration int64, errGM *errorsGM.ErrorGopherMart) {
	get := storage + "/api/orders/" + order
	resp, err := http.Get(get)
	if err != nil {
		return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.UnmarshalError, err}
	}
	switch resp.StatusCode {

	case 200:
		var bodyFloat requestAccrualFloat
		body, err := io.ReadAll(resp.Request.Body)
		if err != nil {
			return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.ReadAllError, err}
		}
		err = json.Unmarshal(body, &bodyFloat)
		if err != nil {
			return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.UnmarshalError, err}
		}
		bodyRequest.Status = bodyFloat.Status
		bodyRequest.Point = uint(bodyFloat.Point * 100)
		bodyRequest.UploadedAt = bodyFloat.UploadedAt
		bodyRequest.Number = bodyFloat.Number
		return bodyRequest, 0, nil

	case 429:
		header := resp.Header
		a := header["Retry-After"][0]
		sec, err := strconv.ParseInt(a, 10, 0)
		if err != nil {
			return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.StatusTooManyRequests, err}
		}
		return requestAccrual{}, sec, nil //

	case 500:
		return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.StatusInternalServerError, err} //
	}

	return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.RespStatusCodeNotMatch, err}
}
