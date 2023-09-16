package entities

import "github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"

type GetWithdrawsResponseBody struct {
	Data []GetWithdrawsResponseBodyData `json:"data"`
}

type GetWithdrawsResponseBodyData struct {
	OrderID     string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func NewGetWithdrawsResponseBody(storeResult models.WithdrawsDataResult) GetWithdrawsResponseBody {
	var result GetWithdrawsResponseBody

	for _, withdraw := range storeResult.Data {
		newUserWithdraw := GetWithdrawsResponseBodyData{
			OrderID:     withdraw.OrderID,
			Sum:         withdraw.Sun,
			ProcessedAt: withdraw.ProcessedAt,
		}
		result.Data = append(result.Data, newUserWithdraw)
	}

	return result
}
