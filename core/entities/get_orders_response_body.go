package entities

import "github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"

type GetOrdersResponseBody struct {
	Data []GetOrdersResponseBodyData `json:"data"`
}

type GetOrdersResponseBodyData struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

func NewGetOrdersResponseBody(data models.GetOrdersDataResult) GetOrdersResponseBody {
	var result GetOrdersResponseBody

	for _, order := range data.Orders {
		responseOrder := GetOrdersResponseBodyData{
			Number:     order.Number,
			Status:     order.Status.String(),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
		result.Data = append(result.Data, responseOrder)
	}

	return result
}
