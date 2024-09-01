package services

import (
	"context"
	"errors"
	"testing"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

// Mock for AuthStore
type MockAuthStore struct {
	users map[string]*models.User
}

func (m *MockAuthStore) SelectUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, exists := m.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockAuthStore) InsertUser(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.Username]; exists {
		return errors.New("user already exists")
	}
	m.users[user.Username] = user
	return nil
}

func (m *MockAuthStore) UpdateUser(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.Username]; !exists {
		return errors.New("user not found")
	}
	m.users[user.Username] = user
	return nil
}

// Helper function to create a new mock store
func NewMockAuthStore() *MockAuthStore {
	return &MockAuthStore{
		users: make(map[string]*models.User),
	}
}

// Test for CreateUser
func TestCreateUser(t *testing.T) {
	mockStore := NewMockAuthStore()
	authService := NewAuthService("secret_key", mockStore)

	// Test user creation
	user, err := authService.CreateUser(context.Background(), "testuser", "password123")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)

	// Test user creation with existing username
	_, err = authService.CreateUser(context.Background(), "testuser", "password123")
	assert.Error(t, err)
	assert.Equal(t, "there's a registered user with this username address", err.Error())
}

// Test for GetUser
func TestGetUser(t *testing.T) {
	mockStore := NewMockAuthStore()
	authService := NewAuthService("secret_key", mockStore)

	// Create a user for testing
	user, _ := authService.CreateUser(context.Background(), "testuser", "password123")

	// Test correct password
	retrievedUser, err := authService.GetUser(context.Background(), "testuser", "password123")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)

	// Test incorrect password
	_, err = authService.GetUser(context.Background(), "testuser", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, "invalid password", err.Error())
}

// Test for ChangePassword
func TestChangePassword(t *testing.T) {
	mockStore := NewMockAuthStore()
	authService := NewAuthService("secret_key", mockStore)

	// Create a user for testing
	_, _ = authService.CreateUser(context.Background(), "testuser", "password123")

	// Test changing password
	err := authService.ChangePassword(context.Background(), "testuser", "newpassword123")
	assert.NoError(t, err)

	// Test login with new password
	_, err = authService.GetUser(context.Background(), "testuser", "newpassword123")
	assert.NoError(t, err)

	// Test login with old password (should fail)
	_, err = authService.GetUser(context.Background(), "testuser", "password123")
	assert.Error(t, err)
	assert.Equal(t, "invalid password", err.Error())
}
