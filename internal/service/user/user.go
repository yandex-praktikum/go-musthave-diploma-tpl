package user

import (
	"context"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	sql "github.com/GTech1256/go-musthave-diploma-tpl/internal/repository/user"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
)

type Storage interface {
	Ping(ctx context.Context) error
	Register(ctx context.Context, userRegister *entity.UserRegisterJSON) (*entity.UserDB, error)
	Login(ctx context.Context, userRegister *entity.UserLoginJSON) (*entity.UserDB, error)
	GetByID(ctx context.Context, userID int) (*entity.UserDB, error)
	Withdraw(ctx context.Context, userID int, withdrawCount float64) (*entity.UserDB, error)
	GetWithdrawals(ctx context.Context, userID int) ([]*entity.WithdrawalDB, error)
	AddWithdrawRecord(ctx context.Context, withdrawalRawRecord entity.WithdrawalRawRecord) error
}

type userService struct {
	logger  logging2.Logger
	storage Storage
	cfg     *config.Config
}

var (
	ErrNotUniqueLogin                      = errors.New("пользователь с таким логином уже зарегистрирован")
	ErrInvalidLoginPasswordCombination     = errors.New("неверная пара логин/пароль")
	ErrWithdrawCountGreaterThanUserBalance = errors.New("запрошенная сумма вывода больше, чем баланс пользователя")
)

func NewUserService(logger logging2.Logger, storage Storage, cfg *config.Config) *userService {
	return &userService{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
	}
}

func (u userService) Ping(ctx context.Context) error {
	return u.storage.Ping(ctx)
}

func (u userService) Register(ctx context.Context, userRegister *entity.UserRegisterJSON) (*entity.UserDB, error) {
	userDB, err := u.storage.Register(ctx, userRegister)
	if errors.Is(err, sql.ErrNotUniqueLogin) {
		return nil, ErrNotUniqueLogin
	}
	if err != nil {
		return nil, err
	}

	return userDB, nil
}

func (u userService) Login(ctx context.Context, userLogin *entity.UserLoginJSON) (*entity.UserDB, error) {
	userDB, err := u.storage.Login(ctx, userLogin)
	if errors.Is(err, sql.ErrInvalidLoginPasswordCombination) {
		return nil, ErrInvalidLoginPasswordCombination
	}
	if err != nil {
		return nil, err
	}

	return userDB, nil
}

func (u userService) GetByID(ctx context.Context, userID int) (*entity.UserDB, error) {
	userDB, err := u.storage.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return userDB, nil
}

func (u userService) GetIsUserExistByIВ(ctx context.Context, userID int) (bool, error) {
	userDB, err := u.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return userDB != nil, nil
}

func (u userService) Withdraw(ctx context.Context, withdrawalRawRecord entity.WithdrawalRawRecord) (*entity.UserDB, error) {
	userDB, err := u.storage.Withdraw(ctx, withdrawalRawRecord.UserID, withdrawalRawRecord.Sum)
	if errors.Is(err, sql.ErrWithdrawCountGreaterThanUserBalance) {
		return nil, ErrWithdrawCountGreaterThanUserBalance
	}
	if err != nil {
		return nil, err
	}

	err = u.storage.AddWithdrawRecord(ctx, withdrawalRawRecord)
	if err != nil {
		return nil, err
	}

	return userDB, nil
}

func (u userService) GetWithdrawals(ctx context.Context, userID int) ([]*entity.WithdrawalDB, error) {
	return u.storage.GetWithdrawals(ctx, userID)
}
