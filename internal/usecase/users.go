package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"

	"golang.org/x/crypto/bcrypt"
)

// GetUserProfile получает профиль пользователя
func (uc *useCase) GetUserProfile(ctx context.Context, userID int) (*entity.User, error) {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Скрываем пароль
	user.Password = ""

	return user, nil
}

// UpdateUserPassword обновляет пароль пользователя
func (uc *useCase) UpdateUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	// Получаем текущего пользователя
	user, err := uc.repo.GetUserByID(ctx, userID)
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
	err = uc.repo.UpdateUserPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeactivateAccount деактивирует аккаунт пользователя
func (uc *useCase) DeactivateAccount(ctx context.Context, userID int) error {
	// Проверяем, что у пользователя нет активных заказов
	orders, err := uc.repo.GetOrdersByUserID(ctx, userID)
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
	err = uc.repo.DeactivateUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	return nil
}
