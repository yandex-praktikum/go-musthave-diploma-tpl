package models

import (
	"encoding/json"
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
