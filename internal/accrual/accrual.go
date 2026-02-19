package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	internalConst "github.com/Raime-34/gophermart.git/internal/accrual/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type AccrualCalculator struct {
	isEmulationMode bool
	baseUrl         string
	client          *resty.Client
	orderStates     map[string]*dto.OrderInfo
	mu              sync.RWMutex
}

func NewAccrualCalculator(accrualServiceUrl string) *AccrualCalculator {
	calculator := AccrualCalculator{
		orderStates: make(map[string]*dto.OrderInfo),
	}

	if accrualServiceUrl == "" {
		calculator.isEmulationMode = true
	} else {
		calculator.baseUrl = composeBaseUrl(accrualServiceUrl)
		calculator.client = resty.New()
	}

	return &calculator
}

// Хэлпер для получения эндпоинта сервиса
func composeBaseUrl(accrualServiceUrl string) string {
	return fmt.Sprintf("%v/api/orders/%%v", accrualServiceUrl)
}

// StartMonitoring запускает фоновый мониторинг заказов.
// Возвращает канал, в который отправляются обновлённые состояния заказов.
func (c *AccrualCalculator) StartMonitoring(ctx context.Context, wg *sync.WaitGroup) <-chan *dto.AccrualCalculatorDTO {
	ch := make(chan *dto.AccrualCalculatorDTO)

	// воркер для опроса внешнего сервиса
	go func(chan<- *dto.AccrualCalculatorDTO) {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Finished order status monitoring")
				close(ch)

				return
			default:
				c.mu.Lock()
				for orderNumber, state := range c.orderStates {
					if state.Status == consts.INVALID || state.Status == consts.PROCESSED {
						continue
					}

					currentStatus, err := c.getOrderState(orderNumber)
					switch err {
					case internalConst.ErrNotRegistered:
						delete(c.orderStates, orderNumber)
						continue
					case internalConst.ErrToManyRequest:
						time.Sleep(1 * time.Minute)
						continue
					default:
						if err != nil {
							continue
						}
					}

					if currentStatus == nil {
						continue
					}
					if !state.IsEqual(currentStatus) {
						state.Update(currentStatus)
						currentStatus.AddUserId(state.GetUserId())
						ch <- currentStatus
					}
				}
				c.mu.Unlock()
			}
		}
	}(ch)

	return ch
}

// Метод, который добавляет заказ для мониторинга
func (c *AccrualCalculator) AddToMonitoring(orderNumber, userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orderStates[orderNumber] = dto.NewOrderInfo(orderNumber, userID)
}

// Хэлпер для гнерации рандомного количества бонусов в режими эмуляции внешнего сервиса
func getRandomAccrual() int {
	return rand.IntN(1000) + 200
}

// getOrderState получает актуальный статус заказа.
// В режиме эмуляции возвращает фиктивные данные.
// В обычном режиме выполняет HTTP-запрос к внешнему сервису.
func (c *AccrualCalculator) getOrderState(orderNumber string) (*dto.AccrualCalculatorDTO, error) {
	if c.isEmulationMode {
		return &dto.AccrualCalculatorDTO{
			Order:   orderNumber,
			Status:  consts.PROCESSED,
			Accrual: getRandomAccrual(),
		}, nil
	}

	res, err := c.client.R().Get(fmt.Sprintf(c.baseUrl, orderNumber))
	if err != nil {
		logger.Error("Failed to getOrderState", zap.Error(err))
		return nil, nil
	}

	switch res.StatusCode() {
	case http.StatusOK:
		dec := json.NewDecoder(res.Body)
		var orderInfo dto.AccrualCalculatorDTO
		err := dec.Decode(&orderInfo)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode")
		}
		return &orderInfo, nil
	case http.StatusNoContent:
		return nil, internalConst.ErrNotRegistered
	case http.StatusTooManyRequests:
		return nil, internalConst.ErrToManyRequest
	case http.StatusInternalServerError:
		return nil, internalConst.ErrInternal
	default:
		return nil, fmt.Errorf("unknown status")
	}
}
