package service

import (
	"context"
	"errors"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// LoyaltyService implements user registration and loyalty operations on top of a Repository and an AuthManager.
type LoyaltyService struct {
	repository  Repository
	authManager AuthManager
	logger      *zap.Logger
}

// NewLoyaltyService creates a service implementing user registration and loyalty operations.
func NewLoyaltyService(repository Repository, authManager AuthManager, logger *zap.Logger) *LoyaltyService {
	return &LoyaltyService{
		repository:  repository,
		authManager: authManager,
		logger:      logger,
	}
}

// RegisterUser validates that the username is free, hashes the password, creates the user and returns a signed token.
func (svc *LoyaltyService) RegisterUser(ctx context.Context, user model.User) (string, error) {
	found, err := svc.repository.UserExists(ctx, user.Username)
	if err != nil {
		return "", err
	}
	if found {
		return "", model.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return "", model.ErrBadPassword
	}
	if err != nil {
		return "", err
	}

	err = svc.repository.CreateUser(ctx, user.Username, string(hashedPassword))
	if err != nil {
		return "", err
	}

	token, err := svc.authManager.IssueToken(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

// LoginUser authenticates the user and returns a signed token.
func (svc *LoyaltyService) LoginUser(ctx context.Context, user model.User) (string, error) {
	hashedPassword, err := svc.repository.GetPasswordHash(ctx, user.Username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", model.ErrMismatchedPassword
	}
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return "", model.ErrBadPassword
	}
	if err != nil {
		return "", err
	}

	token, err := svc.authManager.IssueToken(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

// IdentifyUser returns the username bound to the given token.
func (svc *LoyaltyService) IdentifyUser(token string) (string, error) {
	username, err := svc.authManager.VerifyToken(token)
	if err != nil {
		return "", err
	}
	return username, nil
}

// LoadOrder registers a new order for accrual processing.
func (svc *LoyaltyService) LoadOrder(ctx context.Context, orderNumber string, username string) error {
	err := model.ValidateOrderNumber(orderNumber)
	if err != nil {
		return err
	}

	orderWriteResult, err := svc.repository.RegisterAccrual(ctx, orderNumber, username)
	if err != nil {
		return err
	}

	if orderWriteResult.Inserted {
		return nil
	}

	if orderWriteResult.Owner != username {
		return model.ErrOrderExistsOtherUser
	}

	return model.ErrOrderExistsSameUser
}

// ListOrders returns all orders of the given user.
func (svc *LoyaltyService) ListOrders(ctx context.Context, username string) ([]model.AccrualResponse, error) {
	orders, err := svc.repository.ListAccruals(ctx, username)
	if err != nil {
		return orders, err
	}
	return orders, nil
}

// GetBalance returns the current balance and the total withdrawn of the given user.
func (svc *LoyaltyService) GetBalance(ctx context.Context, username string) (model.BalanceResponse, error) {
	balance, err := svc.repository.GetBalance(ctx, username)
	if err != nil {
		return balance, err
	}
	return balance, nil
}

// Withdraw spends loyalty points on a withdrawal request.
func (svc *LoyaltyService) Withdraw(ctx context.Context, username string, withdrawal model.WithdrawalRequest) error {
	err := model.ValidateOrderNumber(withdrawal.OrderNumber)
	if err != nil {
		return err
	}

	err = svc.repository.Withdraw(ctx, username, withdrawal)
	if err != nil {
		return err
	}

	// audit-log
	svc.logger.Info("withdrawal recorded",
		zap.String("username", username),
		zap.String("order", withdrawal.OrderNumber),
		zap.Float64("sum", withdrawal.Sum),
	)
	return nil
}

// ListWithdrawals returns the withdrawal history of the given user.
func (svc *LoyaltyService) ListWithdrawals(ctx context.Context, username string) ([]model.WithdrawalResponse, error) {
	response, err := svc.repository.ListWithdrawals(ctx, username)
	if err != nil {
		return response, err
	}
	return response, nil
}
