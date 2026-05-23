package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/services"
	"go-musthave-diploma-tpl/internal/storage"
)

// MockUserRepository реализует repository.UserRepository для тестов
type MockUserRepository struct {
	users map[string]*storage.User
	byID  map[int64]*storage.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*storage.User),
		byID:  make(map[int64]*storage.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, login string, passwordHash []byte) (*storage.User, error) {
	if _, exists := m.users[login]; exists {
		return nil, repository.ErrUserAlreadyExists
	}

	id := int64(len(m.users) + 1)
	user := &storage.User{
		ID:           id,
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	m.users[login] = user
	m.byID[id] = user
	return user, nil
}

func (m *MockUserRepository) GetByLogin(ctx context.Context, login string) (*storage.User, error) {
	user, exists := m.users[login]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*storage.User, error) {
	user, exists := m.byID[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

// ============ Тесты регистрации ============

func TestUserService_Register_Success(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()
	login := "testuser"
	password := "testpass123"

	// Act
	token, err := userService.Register(ctx, login, password)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем, что пользователь создан
	user, err := mockRepo.GetByLogin(ctx, login)
	require.NoError(t, err)
	assert.Equal(t, login, user.Login)
}

func TestUserService_Register_DuplicateLogin(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()
	login := "duplicateuser"
	password := "testpass123"

	// Первая регистрация - успех
	_, err := userService.Register(ctx, login, password)
	require.NoError(t, err)

	// Act - вторая регистрация с тем же логином
	_, err = userService.Register(ctx, login, password)

	// Assert
	assert.ErrorIs(t, err, services.ErrLoginAlreadyExists)
}

// ============ Тесты логина ============

func TestUserService_Login_Success(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()
	login := "loginuser"
	password := "testpass123"

	// Создаём пользователя
	_, err := userService.Register(ctx, login, password)
	require.NoError(t, err)

	// Act - логин
	token, err := userService.Login(ctx, login, password)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()
	login := "loginuser"
	password := "testpass123"
	wrongPassword := "wrongpass"

	// Создаём пользователя
	_, err := userService.Register(ctx, login, password)
	require.NoError(t, err)

	// Act - логин с неправильным паролем
	_, err = userService.Login(ctx, login, wrongPassword)

	// Assert
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()
	login := "nonexistent"
	password := "testpass123"

	// Act
	_, err := userService.Login(ctx, login, password)

	// Assert
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

// ============ Тесты получения пользователя ============

func TestUserService_GetUserByID_Success(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()
	login := "testuser"
	password := "testpass123"

	// Создаём пользователя
	_, err := userService.Register(ctx, login, password)
	require.NoError(t, err)

	// Получаем пользователя
	createdUser, err := mockRepo.GetByLogin(ctx, login)
	require.NoError(t, err)

	// Act
	user, err := userService.GetUserByID(ctx, createdUser.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, login, user.Login)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := NewMockUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()

	// Act
	_, err := userService.GetUserByID(ctx, 999)

	// Assert
	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

// ============ Пример с более сложным моком ============

// MockUserRepositoryWithError возвращает ошибку на определённый вызов
type MockUserRepositoryWithError struct {
	failOnCreate bool
	failOnGet    bool
}

func (m *MockUserRepositoryWithError) Create(ctx context.Context, login string, passwordHash []byte) (*storage.User, error) {
	if m.failOnCreate {
		return nil, errors.New("database connection failed")
	}
	return &storage.User{ID: 1, Login: login, PasswordHash: passwordHash}, nil
}

func (m *MockUserRepositoryWithError) GetByLogin(ctx context.Context, login string) (*storage.User, error) {
	if m.failOnGet {
		return nil, errors.New("database connection failed")
	}
	return &storage.User{ID: 1, Login: login}, nil
}

func (m *MockUserRepositoryWithError) GetByID(ctx context.Context, id int64) (*storage.User, error) {
	if m.failOnGet {
		return nil, errors.New("database connection failed")
	}
	return &storage.User{ID: id}, nil
}

func TestUserService_Register_DatabaseError(t *testing.T) {
	// Arrange
	mockRepo := &MockUserRepositoryWithError{failOnCreate: true}
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	userService := services.NewUserService(mockRepo, tokenManager)

	ctx := context.Background()

	// Act
	_, err := userService.Register(ctx, "testuser", "testpass")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
}
