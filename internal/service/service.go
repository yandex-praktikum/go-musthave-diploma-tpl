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
	orderQueue  chan<- string
	logger      *zap.Logger
}

// NewLoyaltyService creates a service implementing user registration and loyalty operations.
func NewLoyaltyService(repository Repository, authManager AuthManager, orderQueue chan<- string, logger *zap.Logger) *LoyaltyService {
	return &LoyaltyService{
		repository:  repository,
		authManager: authManager,
		orderQueue:  orderQueue,
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

	token, err := svc.authManager.IssueToken(user.Username)
	if err != nil {
		return "", err
	}

	err = svc.repository.CreateUser(ctx, user.Username, string(hashedPassword))
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
func (svc *LoyaltyService) LoadOrder(ctx context.Context, orderNumber string, user model.User) error {
	var order = &model.Order{
		Username: user.Username,
		Number:   orderNumber,
	}

	err := order.Validate()
	if err != nil {
		return err
	}

	orderWriteResult, err := svc.repository.RegisterOrder(ctx, orderNumber, user.Username)
	if err != nil {
		return err
	}

	if orderWriteResult.Inserted {
		select {
		case svc.orderQueue <- orderNumber:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	if orderWriteResult.Owner != user.Username {
		return model.ErrOrderExistsOtherUser
	}

	return model.ErrOrderExistsSameUser
}

// ListOrders returns all orders of the given user.
func (svc *LoyaltyService) ListOrders(user model.User) error {
	return nil
}

// GetBalance returns the current balance and total accrued of the given user.
func (svc *LoyaltyService) GetBalance(user model.User) error {
	return nil
}

// Withdraw spends loyalty points on a withdrawal request.
func (svc *LoyaltyService) Withdraw(user model.User) error {
	return nil
}

// ListWithdrawals returns the withdrawal history of the given user.
func (svc *LoyaltyService) ListWithdrawals(user model.User) error {
	return nil
}
