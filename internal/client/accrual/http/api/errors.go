package api

import "errors"

var (
	ErrRequestInitiate = errors.New("при формировании запроса произошла ошибка")
	ErrRequestDo       = errors.New("при запросе произошла ошибка ")
	//ErrInvalidResponseStatus = errors.New("ответ пришел не с 200 статусом ответа")
)
