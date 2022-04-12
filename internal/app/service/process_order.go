package service

import (
	"errors"
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/service/methods"
	"log"
	"time"
)

/*
Доступные статусы обработки заказов:
NEW — заказ загружен в систему, но не попал в обработку. Статус проставляется при первичном попадании в БД.
PROCESSING — вознаграждение за заказ рассчитывается. Статус проставляется при получении статусов REGISTERED & PROCESSING
INVALID — система расчёта вознаграждений отказала в расчёте;
PROCESSED — данные по заказу проверены и информация о расчёте успешно получена.
*/

func ProcessOrder(orderNumber string) error {
	// func to get all orders where the status isn't final? and then run in loop
	orderInfo, err := GetOrderAccrualInfo(orderNumber)
	if err != nil {
		if errors.Is(err, ErrAccSysTooManyReq) {
			time.Sleep(time.Second * 60)
			return ProcessOrder(orderInfo.Order)
		}
		log.Println("[ERROR] accrual system", err)
	}
	ord := methods.NewOrder(orderInfo.Order)

	switch orderInfo.Status {
	case "REGISTERED": // not final status - wait and ask again
		ord.UpdateStatus("PROCESSING")
		time.Sleep(time.Second * 60)
		return ProcessOrder(orderInfo.Order)
	case "PROCESSING": // not final status - wait and ask again
		ord.UpdateStatus("PROCESSING")
		time.Sleep(time.Second * 60)
		return ProcessOrder(orderInfo.Order)
	case "INVALID": // final status - error, update data in table orders
		return ord.UpdateStatus("INVALID")
	case "PROCESSED": // final status - success, update data in table orders
		ord.UpdateStatus("PROCESSED")
		ord.SetAccrual(orderInfo.Accrual)
		bal := methods.NewBalanceRecord()
		bal.OrderNumber = ord.Number
		bal.Income = ord.Accrual
		return bal.Add()
	}
	return nil
}
