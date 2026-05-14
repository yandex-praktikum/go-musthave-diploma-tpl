package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/repository"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Register(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(password string) func(m *MockRepository, mh *MockHasher)
		expectedResult bool
		expectedError  error
	}{
		{
			name: "valid register",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					hashedPassword := "hash"
					mr.On("SaveUser", mock.Anything, mock.Anything, hashedPassword).Return("id", nil)

					mh.On("HashPassword", correctPassword).Return(hashedPassword, nil)
				}
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "user already exists",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					hashedPassword := "hash"
					mr.On("SaveUser", mock.Anything, mock.Anything, hashedPassword).Return("", repository.ErrorLoginAlreadyExists)

					mh.On("HashPassword", correctPassword).Return(hashedPassword, nil)
				}
			},
			expectedResult: false,
			expectedError:  model.ErrorUserAlreadyExists,
		},

		{
			name: "internal service error",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					hashedPassword := "hash"
					mr.On("SaveUser", mock.Anything, mock.Anything, hashedPassword).Return("", fmt.Errorf("error"))

					mh.On("HashPassword", correctPassword).Return(hashedPassword, nil)
				}
			},
			expectedResult: false,
			expectedError:  fmt.Errorf("error"),
		},

		{
			name: "invalid hash result",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					mh.On("HashPassword", correctPassword).Return("", fmt.Errorf("error"))
				}
			},
			expectedResult: false,
			expectedError:  fmt.Errorf("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup("password")(mockRepo, mockHasher)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			result, err := service.Register(context.Background(), &model.LoginPassword{
				Login:    "login",
				Password: "password",
			})

			if tt.expectedResult {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(password string) func(m *MockRepository, mh *MockHasher)
		result        string
		expectedError error
	}{
		{
			name: "valid login",
			mockSetup: func(password string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					hashedPassword := "hash"
					mr.On("GetPasswordByEmail", mock.Anything, mock.Anything).Return("id", hashedPassword, nil)
					mh.On("VerifyPassword", password, hashedPassword).Return(true, nil)
				}
			},
			result:        "id",
			expectedError: nil,
		},
		{
			name: "user doesn't exists",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					mr.On("GetPasswordByEmail", mock.Anything, mock.Anything).Return(mock.Anything, mock.Anything, repository.ErrorNoUser)
				}
			},
			result:        "",
			expectedError: model.ErrorNoUser,
		},

		{
			name: "internal service error",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					mr.On("GetPasswordByEmail", mock.Anything, mock.Anything).Return(mock.Anything, mock.Anything, fmt.Errorf("error"))
				}
			},
			result:        "",
			expectedError: fmt.Errorf("error"),
		},

		{
			name: "invalid password",
			mockSetup: func(correctPassword string) func(mr *MockRepository, mh *MockHasher) {
				return func(mr *MockRepository, mh *MockHasher) {
					hashedPassword := "hash"
					mr.On("GetPasswordByEmail", mock.Anything, mock.Anything).Return(mock.Anything, hashedPassword, nil)

					mh.On("VerifyPassword", correctPassword, hashedPassword).Return(false, nil)
				}
			},
			result:        "",
			expectedError: model.ErrorUnverifiedAccount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup("password")(mockRepo, mockHasher)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			result, err := service.Login(context.Background(), &model.LoginPassword{
				Login:    "login",
				Password: "password",
			})

			assert.Equal(t, tt.result, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ConfirmOrder(t *testing.T) {
	tests := []struct {
		name          string
		order         string
		mockSetup     func(m *MockRepository)
		result        bool
		expectedError error
	}{
		{
			name:  "success",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SaveOrder", mock.Anything, mock.Anything).Return(nil)
			},
			result:        false,
			expectedError: nil,
		},
		{
			name:  "internal error",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SaveOrder", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
			},
			result:        false,
			expectedError: fmt.Errorf("error"),
		},
		{
			name:  "order already exists",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SaveOrder", mock.Anything, mock.Anything).Return(repository.ErrorOrderConflict)
			},
			result:        true,
			expectedError: model.ErrorOrderConflict,
		},
		{
			name:  "order already exists",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SaveOrder", mock.Anything, mock.Anything).Return(repository.ErrorOrderAlreadyExists)
			},
			result:        true,
			expectedError: nil,
		},
		{
			name:          "invalid order",
			order:         "4111111111211111",
			mockSetup:     func(mr *MockRepository) {},
			result:        false,
			expectedError: model.ErrorWrongOrders,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup(mockRepo)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			result, err := service.ConfirmOrder(context.Background(), tt.order)

			assert.Equal(t, tt.result, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetOrdersOfUser(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(m *MockRepository)
		result        bool
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func(mr *MockRepository) {
				mr.On("GetOrders", mock.Anything, mock.Anything).Return([]model.Order{{Number: "1"}}, nil)
			},
			result:        true,
			expectedError: nil,
		},
		{
			name: "internal service error",
			mockSetup: func(mr *MockRepository) {
				mr.On("GetOrders", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
			},
			result:        false,
			expectedError: fmt.Errorf("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup(mockRepo)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			result, err := service.GetOrdersOfUser(context.Background())

			if tt.result {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetBalance(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(m *MockRepository)
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func(mr *MockRepository) {
				mr.On("GetBalance", mock.Anything, mock.Anything).Return(model.Balance{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "internal service error",
			mockSetup: func(mr *MockRepository) {
				mr.On("GetBalance", mock.Anything, mock.Anything).Return(model.Balance{}, fmt.Errorf("error"))
			},
			expectedError: fmt.Errorf("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup(mockRepo)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			_, err := service.GetBalance(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Withdrawn(t *testing.T) {
	tests := []struct {
		name          string
		order         string
		mockSetup     func(m *MockRepository)
		expectedError error
	}{
		{
			name:  "success",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SetWithdrawnOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "internal error",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SetWithdrawnOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
			},
			expectedError: fmt.Errorf("error"),
		},
		{
			name:  "order already exists",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SetWithdrawnOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.ErrorInsufficientBalance)
			},
			expectedError: model.ErrorInsufficientBalance,
		},
		{
			name:  "order already exists",
			order: "49927398716",
			mockSetup: func(mr *MockRepository) {
				mr.On("SetWithdrawnOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.ErrorOrderAlreadyExists)
			},
			expectedError: model.ErrorOrderAlreadyExists,
		},
		{
			name:          "invalid order",
			order:         "4111111111211111",
			mockSetup:     func(mr *MockRepository) {},
			expectedError: model.ErrorWrongOrders,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup(mockRepo)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			err := service.Withdrawn(context.Background(), &model.BalanceWithdraw{
				Order: tt.order,
				Sum:   0.0,
			})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetWithdrawns(t *testing.T) {
	tests := []struct {
		name          string
		result        bool
		mockSetup     func(m *MockRepository)
		expectedError error
	}{
		{
			name:   "success",
			result: true,
			mockSetup: func(mr *MockRepository) {
				mr.On("SetWithdrawnsByUser", mock.Anything, mock.Anything).Return([]model.Withdraw{{Order: "1"}}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "internal service error",
			result: false,
			mockSetup: func(mr *MockRepository) {
				mr.On("SetWithdrawnsByUser", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
			},
			expectedError: fmt.Errorf("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockHasher := new(MockHasher)
			tt.mockSetup(mockRepo)

			service := &Service{
				repo:   mockRepo,
				hasher: mockHasher,
			}

			result, err := service.GetWithdrawns(context.Background())
			if tt.result {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}
func (m *MockHasher) VerifyPassword(password, encodedPassword string) (bool, error) {
	args := m.Called(password, encodedPassword)
	return args.Bool(0), args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveUser(ctx context.Context, login, hash_password string) (string, error) {
	args := m.Called(ctx, login, hash_password)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetPasswordByEmail(ctx context.Context, login string) (string, string, error) {
	args := m.Called(ctx, login)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockRepository) SaveOrder(ctx context.Context, order string) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) GetOrders(ctx context.Context, userID string) ([]model.Order, error) {
	args := m.Called(ctx, userID)

	var orders []model.Order
	if args.Get(0) != nil {
		if val, ok := args.Get(0).([]model.Order); ok {
			orders = val
		}
	}

	return orders, args.Error(1)
}

func (m *MockRepository) GetBalance(ctx context.Context, userID string) (model.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(model.Balance), args.Error(1)
}

func (m *MockRepository) SetWithdrawnOrder(ctx context.Context, userID, orderNumber string, withdrawn float64) error {
	args := m.Called(ctx, userID, orderNumber, withdrawn)
	return args.Error(0)
}

func (m *MockRepository) SetWithdrawnsByUser(ctx context.Context, userID string) ([]model.Withdraw, error) {
	args := m.Called(ctx, userID)

	var withdraw []model.Withdraw
	if args.Get(0) != nil {
		if val, ok := args.Get(0).([]model.Withdraw); ok {
			withdraw = val
		}
	}
	return withdraw, args.Error(1)
}
