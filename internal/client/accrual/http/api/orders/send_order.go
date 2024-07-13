package orders

import (
	"context"
	"errors"
	"fmt"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/dto"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/http/api"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/http/api/orders/converter"
	"net/http"
	"strconv"
)

var (
	ErrUnknownStatusCode            = errors.New("неожиданный ответ сервиса accrual")
	ErrUnknownOrderNumberForAccrual = errors.New("заказ не зарегистрирован в системе расчёта")
	ErrStatusTooManyRequests        = errors.New("превышено количество запросов к сервису")
	ErrStatusInternalServerError    = errors.New("внутренняя ошибка сервера")
)

type TooManyRequestsError struct {
	RetryAfter uint
	Err        error
}

// Error добавляет поддержку интерфейса error для типа TimeError.
func (te *TooManyRequestsError) Error() string {
	return fmt.Sprintf("%v %v", te.RetryAfter, te.Err)
}

// NewTooManyRequestsError записывает ошибку err в тип TimeError c текущим временем.
func NewTooManyRequestsError(err error, retryAfter uint) error {
	return &TooManyRequestsError{
		RetryAfter: retryAfter,
		Err:        err,
	}
}

func (s update) SendOrder(ctx context.Context, orderDTO dto.Order) (*dto.OrderResponse, error) {
	requestURL := getRequestURL(s.BaseURL, &orderDTO)

	req, err := s.HTTPClient.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		s.logger.Errorf("[accrual]: SendOrder - Невозможно создать запрос: %s", err)
		return nil, api.ErrRequestInitiate
	}

	res, err := http.DefaultClient.Do(req) // s.HTTPClient.Do(req)
	if err != nil {
		s.logger.Errorf("[accrual]: SendOrder Ошибка отправки запроса: %v", err)

		return nil, api.ErrRequestDo
	}

	defer res.Body.Close()

	s.logger.Infof("%d %v \n", res.StatusCode, requestURL)

	orderResponse, err := getResponse(res)
	if err != nil {
		return nil, err
	}

	return orderResponse, nil
}

func getRequestURL(baseURL string, updateDto *dto.Order) string {
	return fmt.Sprintf("%v/api/orders/%v", baseURL, updateDto.Number)
}

func getResponse(response *http.Response) (*dto.OrderResponse, error) {
	switch response.StatusCode {
	case http.StatusOK:
		return converter.ResponseBodyToOrderDTO(&response.Body)

	case http.StatusNoContent:
		return nil, ErrUnknownOrderNumberForAccrual

	case http.StatusTooManyRequests:
		atoi, err := strconv.Atoi(response.Header.Get("Retry-After"))
		if err != nil {
			return nil, NewTooManyRequestsError(ErrStatusTooManyRequests, 60)
		}
		return nil, NewTooManyRequestsError(ErrStatusTooManyRequests, uint(atoi))

	case http.StatusInternalServerError:
		return nil, ErrStatusInternalServerError

	default:
		return nil, ErrUnknownStatusCode
	}
}
