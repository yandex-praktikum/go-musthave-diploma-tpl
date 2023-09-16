package entities

import "github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"

type GetUserBalanceResponseBodyData struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func NewGetUserBalanceResponseBodyData(data models.GetUserBalanceDataResult) GetUserBalanceResponseBodyData {
	return GetUserBalanceResponseBodyData{
		Current:   data.Current,
		Withdrawn: data.Withdrawn,
	}
}
