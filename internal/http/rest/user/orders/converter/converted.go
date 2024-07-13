package converter

import (
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"time"
)

func GetOrdersStatusJSONsByOrderDBs(orders []*entity.OrderDB) []*entity.OrderStatusJSON {
	ordersStatusJSONs := make([]*entity.OrderStatusJSON, 0)

	for _, order := range orders {
		ordersStatusJSON := OrderDBToOrderStatusJSON(order)

		ordersStatusJSONs = append(ordersStatusJSONs, ordersStatusJSON)
	}

	return ordersStatusJSONs
}

func OrderDBToOrderStatusJSON(orderDB *entity.OrderDB) *entity.OrderStatusJSON {
	orderStatusJSON := &entity.OrderStatusJSON{
		Number:     orderDB.Number,
		Status:     getOrderStatusByOrderDB(orderDB),
		UploadedAt: orderDB.UploadedAt.Time.Format(time.RFC3339),
	}

	if orderDB.Accrual.Valid {
		orderStatusJSON.Accrual = orderDB.Accrual.Float64
	}

	return orderStatusJSON
}

// Доступные статусы обработки расчётов OrderStatusJSON:
// - `NEW` — заказ загружен в систему, но не попал в обработку;
// - `PROCESSING` — вознаграждение за заказ рассчитывается;
// - `INVALID` — система расчёта вознаграждений отказала в расчёте;
// - `PROCESSED` — данные по заказу проверены и информация о расчёте успешно получена.

// OrderDB.status — статус расчёта начисления:
// - `REGISTERED` — заказ зарегистрирован, но вознаграждение не рассчитано;
// - `INVALID` — заказ не принят к расчёту, и вознаграждение не будет начислено;
// - `PROCESSING` — расчёт начисления в процессе;
// - `PROCESSED` — расчёт начисления окончен;
func getOrderStatusByOrderDB(order *entity.OrderDB) string {
	switch order.Status {
	case "REGISTERED":
		return "NEW"

	case "INVALID":
		return "INVALID"
	case "PROCESSING":
		return "PROCESSING"

	case "PROCESSED":
		return "PROCESSED"

	default:
		return "UNKNOWN"
	}
}
