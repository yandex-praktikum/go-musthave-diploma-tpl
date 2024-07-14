package user

import (
	"context"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
)

type Storage interface {
	Create(ctx context.Context, userId int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
	GetByOrderNumber(ctx context.Context, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
}

type userService struct {
	logger  logging2.Logger
	storage Storage
	cfg     *config.Config
}

var (
	ErrOrderNumberAlreadyUploadByCurrentUser = errors.New("номер заказа уже был загружен этим пользователем")
	ErrOrderNumberAlreadyUploadByOtherUser   = errors.New("номер заказа уже был загружен другим пользователем")
)

func NewOrderService(logger logging2.Logger, storage Storage, cfg *config.Config) *userService {
	return &userService{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
	}
}

func (u userService) Create(ctx context.Context, userId int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error) {
	orderDB, err := u.storage.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if orderDB != nil {
		if orderDB.UserId == userId {
			return nil, ErrOrderNumberAlreadyUploadByCurrentUser
		} else {
			return nil, ErrOrderNumberAlreadyUploadByOtherUser
		}
	}

	orderDB, err = u.storage.Create(ctx, userId, orderNumber)
	if err != nil {
		return nil, err
	}

	return orderDB, nil
}
