package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccrualState string

const (
	// Invalid заказ не принят к расчёту, и вознаграждение не будет начислено
	Invalid AccrualState = "INVALID"
	// New заказ зарегистрирован, но начисление не рассчитано
	New AccrualState = "NEW"
	// Processing расчёт начисления в процессе
	Processing AccrualState = "PROCESSING"
	// Processed расчёт начисления окончен
	Processed AccrualState = "PROCESSED"
)

func StateFromString(state string) AccrualState {
	switch state {
	case string(New):
		return New
	case string(Processing):
		return Processing
	case string(Processed):
		return Processed
	}

	return Invalid
}

type BonusCurrency string

const (
	GopherMarketBonuses = BonusCurrency("GopherMart")
)

type OperationType string

const (
	// Accrual начисление баллов
	Accrual OperationType = "ACCRUAL"
	// Withdraw списание баллов
	Withdraw OperationType = "WITHDRAW"
)

type Orders []Order

type Order struct {
	ID            ID
	UserID        ID
	State         AccrualState
	CreatedAt     time.Time
	AccrualAmount Money
}

type Money struct {
	Currency string
	Amount   decimal.Decimal
}

type Operation struct {
	ID       ID
	OrderID  ID
	Type     OperationType
	Amount   Money
	CratedAt time.Time
}

type Balance struct {
	Accrual   Money
	Withdrawn Money
}

func (b Balance) GetCurrent() Money {
	return Money{
		Amount:   b.Accrual.Amount.Sub(b.Withdrawn.Amount),
		Currency: b.Accrual.Currency,
	}
}
