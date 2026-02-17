// Package helper содержит конвертеры между слоями
package helper

import (
	"github.com/google/uuid"

	"github.com/anon-d/gophermarket/internal/domain"
	"github.com/anon-d/gophermarket/internal/repository"
)

// ToRepositoryUser конвертирует domain.User в repository.User
func ToRepositoryUser(user *domain.User) *repository.User {
	uid, _ := uuid.Parse(user.ID)
	return &repository.User{
		ID:       uid,
		Login:    user.Login,
		PassHash: user.PassHash,
	}
}

// ToDomainUser конвертирует repository.User в domain.User
func ToDomainUser(user *repository.User) *domain.User {
	return &domain.User{
		ID:       user.ID.String(),
		Login:    user.Login,
		PassHash: user.PassHash,
	}
}

// ToRepositoryOrder конвертирует domain.Order в repository.Order
func ToRepositoryOrder(order *domain.Order) *repository.Order {
	uid, _ := uuid.Parse(order.UserID)
	return &repository.Order{
		ID:         order.ID,
		Number:     order.Number,
		UserID:     uid,
		Status:     string(order.Status),
		Accrual:    order.Accrual,
		UploadedAt: order.UploadedAt,
	}
}

// ToDomainOrder конвертирует repository.Order в domain.Order
func ToDomainOrder(order *repository.Order) *domain.Order {
	return &domain.Order{
		ID:         order.ID,
		Number:     order.Number,
		UserID:     order.UserID.String(),
		Status:     domain.OrderStatus(order.Status),
		Accrual:    order.Accrual,
		UploadedAt: order.UploadedAt,
	}
}

// ToDomainOrders конвертирует список repository.Order в список domain.Order
func ToDomainOrders(orders []repository.Order) []domain.Order {
	result := make([]domain.Order, len(orders))
	for i, order := range orders {
		result[i] = *ToDomainOrder(&order)
	}
	return result
}

// ToRepositoryBalance конвертирует domain.Balance в repository.Balance
func ToRepositoryBalance(balance *domain.Balance) *repository.Balance {
	uid, _ := uuid.Parse(balance.UserID)
	return &repository.Balance{
		UserID:    uid,
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}
}

// ToDomainBalance конвертирует repository.Balance в domain.Balance
func ToDomainBalance(balance *repository.Balance) *domain.Balance {
	return &domain.Balance{
		UserID:    balance.UserID.String(),
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}
}

// ToRepositoryWithdrawal конвертирует domain.Withdrawal в repository.Withdrawal
func ToRepositoryWithdrawal(w *domain.Withdrawal) *repository.Withdrawal {
	uid, _ := uuid.Parse(w.UserID)
	return &repository.Withdrawal{
		ID:          w.ID,
		UserID:      uid,
		OrderNumber: w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt,
	}
}

// ToDomainWithdrawal конвертирует repository.Withdrawal в domain.Withdrawal
func ToDomainWithdrawal(w *repository.Withdrawal) *domain.Withdrawal {
	return &domain.Withdrawal{
		ID:          w.ID,
		UserID:      w.UserID.String(),
		OrderNumber: w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt,
	}
}

// ToDomainWithdrawals конвертирует список repository.Withdrawal в список domain.Withdrawal
func ToDomainWithdrawals(withdrawals []repository.Withdrawal) []domain.Withdrawal {
	result := make([]domain.Withdrawal, len(withdrawals))
	for i, w := range withdrawals {
		result[i] = *ToDomainWithdrawal(&w)
	}
	return result
}
