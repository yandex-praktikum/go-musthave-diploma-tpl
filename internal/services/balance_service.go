//go:generate mockgen -source=balance_service.go -destination=mocks/mock_balance_service.go -package=mocks
package services

import (
	"context"

	repository "github.com/KalessinD/gophermart/internal/repositories/postgresql"
)

type (
	/*
		Объект службы балансов
	*/
	BalanceActions struct {
		db repository.SQLStorageInterface
	}

	/*
		Интерфейс оОъекта службы балансов
	*/
	BalanceActionsInterface interface {
		Store(ctx context.Context) error
	}
)

/*
Конструктор службы балансов
*/
func NewBalanceActions(db repository.SQLStorageInterface) OrderActionsInterface {
	return &OrderActions{
		db: db,
	}
}

/*
Сохраняет баланс
*/
func (s *BalanceActions) Store(_ context.Context) error {
	_ = s.db
	return nil
}
