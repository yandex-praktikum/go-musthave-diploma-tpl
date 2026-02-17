package dto

import (
	"time"

	"github.com/Raime-34/gophermart.git/internal/consts"
)

type OrderInfo struct {
	Number     string    `json: "number"`
	Status     string    `json: "status"`
	Accrual    int       `json: "accrual"`
	UploadedAt time.Time `json: "uploaded_at"`

	userID string
}

func NewOrderInfo(orderNumber, userID string) *OrderInfo {
	return &OrderInfo{
		Number:     orderNumber,
		Status:     consts.REGISTERED,
		Accrual:    0,
		UploadedAt: time.Now(),

		userID: userID,
	}
}

func (i *OrderInfo) ToGetOrdersInfoResp() *GetOrdersInfoResp {
	return &GetOrdersInfoResp{
		Number:     i.Number,
		Status:     i.Status,
		UploadedAt: i.UploadedAt,
	}
}

func (i *OrderInfo) IsEqual(other *AccrualCalculatorDTO) bool {
	return i.Status == other.Status
}

func (i *OrderInfo) Update(other *AccrualCalculatorDTO) {
	i.Status = other.Status
	i.Accrual = other.Accrual
}

func (i *OrderInfo) GetUserId() string {
	return i.userID
}

type BalanceInfo struct {
	Current  float64 `json: "current"`
	Withdraw int     `json: "withdrawn"`
}

type GetOrdersInfoResp struct {
	Number     string    `json: "number"`
	Status     string    `json: "status"`
	UploadedAt time.Time `json: "uploaded_at"`
}
