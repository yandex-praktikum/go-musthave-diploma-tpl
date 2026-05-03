package models

import (
	"encoding/json"
	"errors"
	"time"
)

type (
	Withdrawn struct {
		ID          string    `json:"-"`
		UserID      string    `json:"-"`
		OrderID     string    `json:"order"`
		Sum         Accrual   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}

	WithdrawnList []*Withdrawn
)

var ErrWithdrawnNotFound = errors.New("withdrawns not found")

/*
Сериализует объект списания баллов в строку JSON
Может вернуть ошибку, коли та случится.
*/
func (m *Withdrawn) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
