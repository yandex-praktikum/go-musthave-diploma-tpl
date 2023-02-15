package events

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"GopherMart/internal/errorsGM"
)

type requestAccrualFloat struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type requestAccrual struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual uint   `json:"accrual"`
}

func AccrualGet(storage string, order string) (bodyUint requestAccrual, duration int64, err error) {
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
			return requestAccrual{}, 0, errorsGM.ErrAccrualGetError
		}
		err = json.Unmarshal(body, &bodyFloat)
		if err != nil {
			return requestAccrual{}, 0, errorsGM.ErrAccrualGetError
		}
		bodyUint.Status = bodyFloat.Status
		bodyUint.Accrual = uint(bodyFloat.Accrual * 100)
		bodyUint.Order = bodyFloat.Order
		return bodyUint, 0, nil

	case 429:
		header := resp.Header
		a := header["Retry-After"][0]
		sec, err := strconv.ParseInt(a, 10, 0)
		if err != nil {
			return requestAccrual{}, 0, &errorsGM.ErrorGopherMart{errorsGM.StatusTooManyRequests, err}
		}
		return requestAccrual{}, sec, nil //

	case 500:
		return requestAccrual{}, 0, errorsGM.ErrAccrualGetError //
	}

	return requestAccrual{}, 0, errorsGM.ErrAccrualGetError
}
