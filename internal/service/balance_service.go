package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gophermart/internal/repository"
)

type Balance struct {
	Current   float64
	Withdrawn float64
}

type BalanceService struct {
	balanceRepo    *repository.BalanceRepository
	withdrawalRepo *repository.WithdrawalRepository
	orderRepo      *repository.OrderRepository
	db             *sql.DB
}

func NewBalanceService(
	balanceRepo *repository.BalanceRepository,
	withdrawalRepo *repository.WithdrawalRepository,
	orderRepo *repository.OrderRepository,
	db *sql.DB,
) *BalanceService {
	return &BalanceService{
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
		orderRepo:      orderRepo,
		db:             db,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (Balance, error) {
	accrued, err := s.balanceRepo.GetAccrued(ctx, userID)
	if err != nil {
		return Balance{}, fmt.Errorf("get accrued: %w", err)
	}

	withdrawn, err := s.balanceRepo.GetWithdrawn(ctx, userID)
	if err != nil {
		return Balance{}, fmt.Errorf("get withdrawn: %w", err)
	}

	return Balance{
		Current:   accrued - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	if order == "" || sum <= 0 {
		return ErrInvalidInput
	}

	if !IsValidOrderNumber(order) {
		return ErrInvalidOrderNumber
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	accrued, err := s.balanceRepo.GetAccruedInTx(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("get accrued: %w", err)
	}

	withdrawn, err := s.balanceRepo.GetWithdrawnInTx(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("get withdrawn: %w", err)
	}

	current := accrued - withdrawn
	if current < sum {
		return ErrInsufficientFunds
	}

	if err := s.withdrawalRepo.CreateInTx(ctx, tx, userID, order, sum, time.Now()); err != nil {
		return fmt.Errorf("create withdrawal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *BalanceService) ListWithdrawals(ctx context.Context, userID int64) ([]repository.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}
	return withdrawals, nil
}

var ErrInsufficientFunds = errors.New("insufficient funds")
