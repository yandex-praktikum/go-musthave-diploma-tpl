package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

// userUC реализация UserUseCase
type userUC struct {
	repo     *repository.Repository
	hashCost int
}

// NewUserUsecase создает новый экземпляр userUC
func NewUserUsecase(repo *repository.Repository, hashCost int) UserUseCase {
	return &userUC{
		repo:     repo,
		hashCost: hashCost,
	}
}

// GetUserProfile получает профиль пользователя
func (uc *userUC) GetUserProfile(ctx context.Context, userID int) (*entity.User, error) {
	user, err := uc.repo.User().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Скрываем пароль для безопасности
	user.Password = ""

	return user, nil
}

// GetUserByLogin получает пользователя по логину
func (uc *userUC) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	if err := validateLogin(login); err != nil {
		return nil, fmt.Errorf("invalid login: %w", err)
	}

	user, err := uc.repo.User().GetByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUserPassword обновляет пароль пользователя
func (uc *userUC) UpdateUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	// Получаем текущего пользователя
	user, err := uc.repo.User().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем старый пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Валидация нового пароля
	if err := validatePassword(newPassword); err != nil {
		return fmt.Errorf("invalid new password: %w", err)
	}

	// Хеширование нового пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), uc.hashCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль
	err = uc.repo.User().UpdatePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeactivateAccount деактивирует аккаунт пользователя
func (uc *userUC) DeactivateAccount(ctx context.Context, userID int) error {
	// Проверяем, что у пользователя нет активных заказов
	orders, err := uc.repo.Order().GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user orders: %w", err)
	}

	// Проверяем наличие заказов в обработке
	for _, order := range orders {
		if order.Status == entity.OrderStatusNew || order.Status == entity.OrderStatusProcessing {
			return errors.New("cannot deactivate account with pending orders")
		}
	}

	// Деактивируем аккаунт
	err = uc.repo.User().Deactivate(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	return nil
}
