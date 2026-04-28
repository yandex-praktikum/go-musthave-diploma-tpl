package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KalessinD/gophermart/internal/handlers"
	"github.com/KalessinD/gophermart/internal/middleware"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/services"
	"github.com/KalessinD/gophermart/internal/services/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const testSecret = "test-secret-key"

func TestCommonHandler_Register(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		body       string
		prepare    func(m *mocks.MockUserCommonActions)
		wantStatus int
	}{
		{
			name: "Successful registration",
			body: `{"login": "testuser", "password": "password123"}`,
			prepare: func(m *mocks.MockUserCommonActions) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "User already exists (conflict)",
			body: `{"login": "testuser", "password": "password123"}`,
			prepare: func(m *mocks.MockUserCommonActions) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(models.ErrUserExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "Internal server error on register",
			body: `{"login": "testuser", "password": "password123"}`,
			prepare: func(m *mocks.MockUserCommonActions) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(errors.New("db connection lost"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Invalid JSON body",
			body:       `{invalid json}`,
			prepare:    func(_ *mocks.MockUserCommonActions) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authService := services.NewAuthService(testSecret)
			mockService := mocks.NewMockUserCommonActions(ctrl)
			if tt.prepare != nil {
				tt.prepare(mockService)
			}

			h := handlers.NewCommonHandler(mockService, authService)
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			ctx := middleware.AddLoggerToContext(req.Context(), logger)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}

			// Проверяем куку при успехе
			if tt.wantStatus == http.StatusOK {
				cookies := rec.Result().Cookies()
				found := false
				for _, c := range cookies {
					// Убрали строгую проверку на "valid-token", проверяем просто наличие значения
					if c.Name == "token" && c.Value != "" {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected auth cookie 'token' to be set")
				}
			}
		})
	}
}

func TestCommonHandler_Login(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		body       string
		prepare    func(m *mocks.MockUserCommonActions)
		wantStatus int
	}{
		{
			name: "Successful login",
			body: `{"login": "testuser", "password": "password123"}`,
			prepare: func(m *mocks.MockUserCommonActions) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Wrong password",
			body: `{"login": "testuser", "password": "wrongpass"}`,
			prepare: func(m *mocks.MockUserCommonActions) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(models.ErrWrongPassword)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "User not found",
			body: `{"login": "nonexistent", "password": "password123"}`,
			prepare: func(m *mocks.MockUserCommonActions) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(models.ErrUserNotFound)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Bad content type",
			body:       `{"login": "test", "password": "test"}`,
			prepare:    func(_ *mocks.MockUserCommonActions) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authService := services.NewAuthService(testSecret)
			mockService := mocks.NewMockUserCommonActions(ctrl)
			if tt.prepare != nil {
				tt.prepare(mockService)
			}

			h := handlers.NewCommonHandler(mockService, authService)

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.body))

			if tt.name != "Bad content type" {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", "text/plain")
			}

			ctx := middleware.AddLoggerToContext(req.Context(), logger)
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
