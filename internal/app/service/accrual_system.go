package service

import (
	"encoding/json"
	"errors"
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/cfg"
	resp "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/router/responses"
	"io"
	"net/http"
)

// GetOrderAccrualInfo is a method, that makes a request GET /api/orders/{orderNumber} to the Accrual System
// As response receives json
// {
//      "order": "<number>",
//      "status": "PROCESSED",
//      "accrual": 500
//  }
/*
Возможные коды ответа:
200 — успешная обработка запроса.
Формат ответа:
  200 OK HTTP/1.1
  Content-Type: application/json
  {
      "order": "<number>",		order — номер заказа
      "status": "PROCESSED", 	status — статус расчёта начисления
      "accrual": 500			accrual — рассчитанные баллы к начислению, при отсутствии начисления — поле отсутствует в ответе
  }

Status:
REGISTERED — заказ зарегистрирован, но не начисление не рассчитано;
INVALID — заказ не принят к расчёту, и вознаграждение не будет начислено;
PROCESSING — расчёт начисления в процессе;
PROCESSED — расчёт начисления окончен;

Errors:
429 — превышено количество запросов к сервису.
Формат ответа:
  429 Too Many Requests HTTP/1.1
  Content-Type: text/plain
  Retry-After: 60

  No more than N requests per minute allowed

500 — внутренняя ошибка сервера.
Заказ может быть взят в расчёт в любой момент после его совершения. Время выполнения расчёта системой не регламентировано
Статусы INVALID и PROCESSED являются окончательными.
Общее количество запросов информации о начислении не ограничено.
*/

var (
	ErrAccSysTooManyReq = errors.New("accrual system too many requests")
	ErrAccSysInternal   = errors.New("accrual system internal error")
)

func GetOrderAccrualInfo(orderNumber string) (*resp.OrderAccrualInfo, error) {
	client := http.Client{}
	accSysPath := cfg.Envs.AccrualSysAddr + "/api/orders/" + orderNumber
	getReq, err := http.NewRequest(http.MethodGet, accSysPath, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(getReq)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusTooManyRequests:
		return nil, ErrAccSysTooManyReq
	case http.StatusInternalServerError:
		return nil, ErrAccSysInternal
	}

	orderAccInf := new(resp.OrderAccrualInfo)
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, orderAccInf)
	if err != nil {
		return nil, err
	}

	return orderAccInf, nil
}
