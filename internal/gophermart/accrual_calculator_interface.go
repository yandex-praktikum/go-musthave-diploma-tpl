package gophermart

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/dto"
)

type accrualCalculator interface {
	StartMonitoring(context.Context) <-chan *dto.AccrualCalculatorDTO
	AddToMonitoring(string)
}
