//go:generate mockgen -source=user_common_service.go -destination=mocks/mock_user_common_service.go -package=mocks
package services

import (
	"context"

	"github.com/KalessinD/gophermart/internal/models"
	repository "github.com/KalessinD/gophermart/internal/repositories/postgresql"

	"golang.org/x/crypto/bcrypt"
)

type (
	/*
		Объект службы для действий пользователя, нетребующих авторизации
	*/
	CommonActions struct {
		db repository.SQLStorageInterface
	}

	/*
		Интерфейс оОъекта службы для действий пользователя, нетребующих авторизации
	*/
	CommonActionsInterface interface {
		Login(ctx context.Context, user *models.User) error
		Register(ctx context.Context, user *models.User) error
	}
)

/*
Конструктор службы для операций нетребующих авторизации пользователя.
*/
func NewCommonAction(db repository.SQLStorageInterface) CommonActionsInterface {
	return &CommonActions{
		db: db,
	}
}

/*
Выполняет вход в систему.
Может вернуть ошибку, если таковая произошла
*/
func (s *CommonActions) Login(ctx context.Context, userFromRequest *models.User) error {
	if err := userFromRequest.Validate(); err != nil {
		return err
	}

	userFromDb, err := s.db.GetUser(ctx, userFromRequest.Login)
	switch {
	case err != nil:
		return err
	case userFromDb == nil:
		return models.ErrUserNotFound
	}

	// bcrypt.CompareHashAndPassword сравнивает открытый пароль и хеш.
	err = bcrypt.CompareHashAndPassword([]byte(userFromDb.Hash), []byte(userFromRequest.Password))
	if err != nil {
		return models.ErrWrongPassword
	}

	return nil
}

/*
Выполняет регистрацию новоого пользователя в системе.
Может вернуть ошибку, если таковая произошла
*/
func (s *CommonActions) Register(ctx context.Context, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if err := s.fillUserIfRequired(user); err != nil {
		return err
	}
	return s.db.AddUser(ctx, user)
}

func (s *CommonActions) fillUserIfRequired(user *models.User) error {
	if user.Password != "" && user.Hash == "" {
		hash, err := s.hashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Hash = hash
	}
	return nil
}

func (s *CommonActions) hashPassword(password string) (string, error) {
	// Cost 10 — это баланс между безопасностью и скоростью.
	// Чем выше число, тем медленнее считается хеш, тем сложнее подобрать пароль.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
