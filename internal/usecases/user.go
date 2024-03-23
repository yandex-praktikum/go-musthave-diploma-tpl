package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"github.com/A-Kuklin/gophermart/internal/storage"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

type UserCreateArgs struct {
	Login    string
	Password []byte
}

type UserUseCases struct {
	strg   storage.UserStorage
	logger logrus.FieldLogger
}

func NewUserUseCases(strg storage.UserStorage, logger logrus.FieldLogger) *UserUseCases {
	return &UserUseCases{
		strg:   strg,
		logger: logger,
	}
}

func (u *UserUseCases) Create(ctx context.Context, args UserCreateArgs) (*entities.User, error) {
	user, err := getUserFromReq(args)
	if err != nil {
		return nil, err
	}

	usr, err := u.strg.CreateUser(ctx, user)
	if err != nil {
		u.logger.WithError(err).Info("CreateUser error")
		return nil, fmt.Errorf("CreateUser error: %w", err)
	}

	return usr, nil
}

func (u *UserUseCases) Login(ctx context.Context, args UserCreateArgs) (*entities.User, error) {
	user, err := getUserFromReq(args)
	if err != nil {
		return nil, err
	}

	usr, err := u.strg.GetUserByLogin(ctx, user)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, fmt.Errorf("%w: login %s is missing", err, user.Login)
	case err != nil:
		u.logger.WithError(err).Info("GetUserByLogin error")
		return nil, fmt.Errorf("GetUserByLogin error: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword(usr.PassHash, args.Password); err != nil {
		u.logger.WithError(err).Infof("Invalid password [%s]: %s", user.Login, err)
		return nil, ErrInvalidPassword
	}

	return usr, nil
}

func getUserFromReq(args UserCreateArgs) (*entities.User, error) {
	hash, err := bcrypt.GenerateFromPassword(args.Password, bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := entities.User{
		Login:    args.Login,
		PassHash: hash,
	}

	return &user, nil
}
