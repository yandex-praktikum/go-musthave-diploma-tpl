package gophermart

import (
	"context"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/dto"
)

//go:generate mockgen -source=accrual_calculator_interface.go -destination=../../mocks/accrual_calculator.go -package=mock
type accrualCalculator interface {
	StartMonitoring(context.Context, *sync.WaitGroup) <-chan *dto.AccrualCalculatorDTO
	AddToMonitoring(string, string)
}
