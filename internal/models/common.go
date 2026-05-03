package models

import (
	"encoding/json"
	"math"
)

type (
	Accrual int // будем хранить целые копейки
)

/*
Возвращает объект модели из JSON
Может вернуть ошибку, коли та случится.
*/
func FromJSON[T any](str []byte) (*T, error) {
	var item T
	err := json.Unmarshal(str, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler.
func (a *Accrual) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = 0
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}

	// Переводим в копейки, округляя до ближайшего целого
	// math.Round нужен для корректной конвертации 5.005 -> 501 копейка
	*a = Accrual(math.Round(f * 100))
	return nil
}
