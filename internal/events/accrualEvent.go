package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

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
		time.Sleep(3 * time.Second)
		fmt.Println("=====AccrualGet==1=== ")
		var bodyFloat requestAccrualFloat
		fmt.Println("=====AccrualGet==1==2= ")
		body, err := io.ReadAll(resp.Request.Body)
		fmt.Println("=====AccrualGet==2=11== ")
		if err != nil {
			fmt.Println("=====AccrualGet==2=== ")
			return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
		}
		fmt.Println("=====AccrualGet==2==1= ")
		err = json.Unmarshal(body, &bodyFloat)
		if err != nil {
			fmt.Println("=====AccrualGet==3=== ")
			return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
		}
		fmt.Println("=====AccrualGet==4=== ")
		bodyUint.Status = bodyFloat.Status
		bodyUint.Accrual = uint(bodyFloat.Accrual * 100)
		bodyUint.Order = bodyFloat.Order
		fmt.Println("=====AccrualGet==5=== ")
		return bodyUint, 0, nil

	case 429:
		fmt.Println("=====AccrualGet===4== ")
		header := resp.Header
		a := header["Retry-After"][0]
		sec, err := strconv.ParseInt(a, 10, 0)
		if err != nil {
			return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
		}
		return requestAccrual{}, sec, nil //

	case 500:
		fmt.Println("=====AccrualGet===5== ")
		return requestAccrual{}, 0, errorsgm.ErrAccrualGetError //
	}

	return requestAccrual{}, 0, errorsgm.ErrAccrualGetError
}
