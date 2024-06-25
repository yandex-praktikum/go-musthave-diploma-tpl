package models

type CreateWithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}
