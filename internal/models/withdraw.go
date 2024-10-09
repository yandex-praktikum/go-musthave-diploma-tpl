package models

type Withdraw struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}
