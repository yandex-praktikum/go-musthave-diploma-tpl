package service

import (
	"context"
	"errors"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/repository"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/utils"
)

var _ IRepo = &repository.Repository{}

type IRepo interface {
	SaveUser(ctx context.Context, login, hashPassword string) (string, error)
	GetPasswordByEmail(ctx context.Context, login string) (string, string, error)

	SaveOrder(ctx context.Context, order string) error
	GetOrders(ctx context.Context, userID string) ([]model.Order, error)
	GetBalance(ctx context.Context, userID string) (model.Balance, error)
	SetWithdrawnOrder(ctx context.Context, userID, orderNumber string, withdrawn float64) error
	SetWithdrawnsByUser(ctx context.Context, userID string) ([]model.Withdraw, error)
}

type iHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, encodedPassword string) (bool, error)
}

type Service struct {
	repo IRepo

	hasher iHasher
}

func NewService(cfg *Config, repo IRepo) *Service {
	hasher := utils.NewHasher(cfg.SaltLen, cfg.KeyLength)
	return &Service{
		repo: repo,

		hasher: hasher,
	}
}

func (s *Service) Register(ctx context.Context, reg *model.LoginPassword) (string, error) {
	hashedPassword, err := s.hasher.HashPassword(reg.Password)
	if err != nil {
		return "", err
	}
	id, err := s.repo.SaveUser(ctx, reg.Login, hashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrorLoginAlreadyExists) {
			return "", model.ErrorUserAlreadyExists
		}
		return "", err
	}
	return id, nil
}

func (s *Service) Login(ctx context.Context, reg *model.LoginPassword) (string, error) {
	id, rightHash, err := s.repo.GetPasswordByEmail(ctx, reg.Login)
	if err != nil {
		if errors.Is(err, repository.ErrorNoUser) {
			return "", model.ErrorNoUser
		}
		return "", err
	}

	verified, err := s.hasher.VerifyPassword(reg.Password, rightHash)
	if err != nil {
		return "", err
	}

	if !verified {
		return "", model.ErrorUnverifiedAccount
	}

	return id, nil
}

func (s *Service) ConfirmOrder(ctx context.Context, order string) (bool, error) {
	if !utils.MoonAlgorithm(order) {
		return false, model.ErrorWrongOrders
	}

	err := s.repo.SaveOrder(ctx, order)
	if err != nil {
		if errors.Is(err, repository.ErrorOrderAlreadyExists) {
			return true, nil
		}
		if errors.Is(err, repository.ErrorOrderConflict) {
			return true, model.ErrorOrderConflict
		}
		return false, err
	}
	return false, nil
}

func (s *Service) GetOrdersOfUser(ctx context.Context) ([]model.Order, error) {
	userID := utils.GetUserID(ctx)
	return s.repo.GetOrders(ctx, userID)
}

func (s *Service) GetBalance(ctx context.Context) (model.Balance, error) {
	userID := utils.GetUserID(ctx)
	return s.repo.GetBalance(ctx, userID)
}

func (s *Service) Withdrawn(ctx context.Context, req *model.BalanceWithdraw) error {
	if !utils.MoonAlgorithm(req.Order) {
		return model.ErrorWrongOrders
	}

	userID := utils.GetUserID(ctx)
	if err := s.repo.SetWithdrawnOrder(ctx, userID, req.Order, req.Sum); err != nil {
		if errors.Is(err, repository.ErrorInsufficientBalance) {
			return model.ErrorInsufficientBalance
		}
		if errors.Is(err, repository.ErrorOrderAlreadyExists) {
			return model.ErrorOrderAlreadyExists
		}
		return err
	}

	return nil
}

func (s *Service) GetWithdrawns(ctx context.Context) ([]model.Withdraw, error) {
	userID := utils.GetUserID(ctx)
	return s.repo.SetWithdrawnsByUser(ctx, userID)
}
