package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

func TestLoginUser_Success(t *testing.T) {
	ctrl, mockService, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	mockService.EXPECT().AuthenticateUser(gomock.Any(), "user1", "password1").Return(nil)

	h.LoginUser(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}
}

func TestLoginUser_InvalidMethod(t *testing.T) {
	ctrl, _, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/user/login", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	h.LoginUser(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestLoginUser_InvalidContentType(t *testing.T) {
	ctrl, _, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "text/plain")

	rr := httptest.NewRecorder()

	h.LoginUser(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}

func TestLoginUser_InvalidJSONPayload(t *testing.T) {
	ctrl, _, h := setup(t)
	defer ctrl.Finish()

	req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	h.LoginUser(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}

func TestLoginUser_NoSuchUserError(t *testing.T) {
	ctrl, mockService, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	mockService.EXPECT().AuthenticateUser(gomock.Any(), "user1", "password1").Return(customerror.ErrNoSuchUser)

	h.LoginUser(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Expected status code %v, got %v", http.StatusUnauthorized, status)
	}
}

func TestLoginUser_ServiceOtherError(t *testing.T) {
	ctrl, mockService, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	mockService.EXPECT().AuthenticateUser(gomock.Any(), "user1", "password1").Return(errors.New("service error"))

	h.LoginUser(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}
