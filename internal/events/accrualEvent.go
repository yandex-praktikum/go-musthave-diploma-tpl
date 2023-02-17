package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"GopherMart/internal/errorsgm"
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
		return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
	}
	switch resp.StatusCode {

	case 200:
		var bodyFloat requestAccrualFloat
		fmt.Println("=====AccrualGet==1=== ")
		fmt.Println(resp)
		fmt.Println("=====AccrualGet==1=1== ")
		fmt.Println(resp.Body)
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		fmt.Println("=====AccrualGet==2=== ")
		if err != nil {
			fmt.Println("=====AccrualGet==3=== ")
			return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
		}
		err = json.Unmarshal(body, &bodyFloat)
		if err != nil {
			return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
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
			return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
		}
		return requestAccrual{}, sec, nil //

	case 500:
		return requestAccrual{}, 0, errorsgm.ErrAccrualGetError //
	}

	return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
}
