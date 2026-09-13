package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/b602op/go-musthave-diploma-tpl/internal/repository"
	"github.com/b602op/go-musthave-diploma-tpl/internal/utils"
)

// BalanceService — сервис получения баланса и списания баллов.
type BalanceService struct {
	balanceRepo    *repository.BalanceRepository
	withdrawalRepo *repository.WithdrawalRepository
	db             *sql.DB
}

// NewBalanceService создаёт новый BalanceService с указанными репозиториями
// балансов и списаний.
func NewBalanceService(
	balanceRepo *repository.BalanceRepository,
	withdrawalRepo *repository.WithdrawalRepository,
	db *sql.DB,
) *BalanceService {
	return &BalanceService{
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
		db:             db,
	}
}

// GetBalance возвращает баланс пользователя
func (s *BalanceService) GetBalance(userID int64) (*domain.Balance, error) {
	balance, err := s.balanceRepo.GetBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// Withdraw списывает баллы со счета пользователя.
//
// Списание баланса и создание записи о списании выполняются в одной
// транзакции: если создание записи падает, баланс откатывается, и баллы
// не теряются.
func (s *BalanceService) Withdraw(userID int64, orderNumber string, sum float64) error {
	if !utils.ValidLuhn(orderNumber) {
		return domain.ErrInvalidOrderNumber
	}

	if sum <= 0 {
		return fmt.Errorf("withdrawal sum must be positive")
	}

	// Открываем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // безопасно: если Commit уже был, Rollback вернёт ErrTxDone

	// 1. Списываем баллы
	if err := s.balanceRepo.SubtractBalanceTx(tx, userID, sum); err != nil {
		return err
	}

	// 2. Создаём запись о списании
	withdrawal := &domain.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}
	if err := s.withdrawalRepo.CreateTx(tx, withdrawal); err != nil {
		return fmt.Errorf("failed to create withdrawal record: %w", err)
	}

	// 3. Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetWithdrawals возвращает историю списаний пользователя
func (s *BalanceService) GetWithdrawals(userID int64) ([]*domain.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	return withdrawals, nil
}
