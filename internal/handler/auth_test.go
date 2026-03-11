package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/config"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/handler"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository/mock"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/router"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setupRouter(t *testing.T, ctrl *gomock.Controller) (*gin.Engine, *mock.MockUserRepository) {
	t.Helper()
	repo := mock.NewMockUserRepository(ctrl)
	services := &service.Services{User: service.NewUserService(repo), Order: nil}
	h := handler.New(services, "test-secret")
	return router.SetupRouter(h, &config.Config{}), repo
}

func TestHandler_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, repo := setupRouter(t, ctrl)
	user := &models.User{
		ID:           1,
		Login:        "alice",
		PasswordHash: "hash",
		Active:       true,
		CreatedAt:    time.Now(),
	}
	repo.EXPECT().
		Create(gomock.Any(), "alice", gomock.Any()).
		Return(user, nil)

	body, _ := json.Marshal(handler.RegisterRequest{Login: "alice", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Register: status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _ := setupRouter(t, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Register invalid JSON: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Register_EmptyLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _ := setupRouter(t, ctrl)

	body, _ := json.Marshal(handler.RegisterRequest{Login: "", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Register empty login: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Register_DuplicateLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, repo := setupRouter(t, ctrl)
	repo.EXPECT().
		Create(gomock.Any(), "bob", gomock.Any()).
		Return(nil, &repository.ErrDuplicateLogin{Login: "bob"})

	body, _ := json.Marshal(handler.RegisterRequest{Login: "bob", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Register duplicate: status = %d; want %d", w.Code, http.StatusConflict)
	}
}

func TestHandler_Register_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, repo := setupRouter(t, ctrl)
	repo.EXPECT().
		Create(gomock.Any(), "user", gomock.Any()).
		Return(nil, http.ErrHandlerTimeout)

	body, _ := json.Marshal(handler.RegisterRequest{Login: "user", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Register internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:           1,
		Login:        "alice",
		PasswordHash: string(hash),
		Active:       true,
		CreatedAt:    time.Now(),
	}

	r, repo := setupRouter(t, ctrl)
	repo.EXPECT().
		GetByLogin(gomock.Any(), "alice").
		Return(user, nil)

	body, _ := json.Marshal(handler.LoginRequest{Login: "alice", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Login: status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _ := setupRouter(t, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Login invalid JSON: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Login_EmptyLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, _ := setupRouter(t, ctrl)

	body, _ := json.Marshal(handler.LoginRequest{Login: "", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Login empty login: status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Login_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, repo := setupRouter(t, ctrl)
	repo.EXPECT().
		GetByLogin(gomock.Any(), "nobody").
		Return(nil, &repository.ErrUserNotFound{Login: "nobody"})

	body, _ := json.Marshal(handler.LoginRequest{Login: "nobody", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Login user not found: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.DefaultCost)
	user := &models.User{
		ID:           1,
		Login:        "alice",
		PasswordHash: string(hash),
		Active:       true,
		CreatedAt:    time.Now(),
	}

	r, repo := setupRouter(t, ctrl)
	repo.EXPECT().
		GetByLogin(gomock.Any(), "alice").
		Return(user, nil)

	body, _ := json.Marshal(handler.LoginRequest{Login: "alice", Password: "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Login wrong password: status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_Login_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, repo := setupRouter(t, ctrl)
	repo.EXPECT().
		GetByLogin(gomock.Any(), "alice").
		Return(nil, http.ErrHandlerTimeout)

	body, _ := json.Marshal(handler.LoginRequest{Login: "alice", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Login internal: status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}
