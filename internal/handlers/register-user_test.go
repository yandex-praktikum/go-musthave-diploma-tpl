package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

func TestRegisterUser_Success(t *testing.T) {
	ctrl, mockService, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	mockService.EXPECT().RegisterUser(gomock.Any(), "user1", "password1").Return(nil)

	h.RegisterUser(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}
}

func TestRegisterUser_InvalidMethod(t *testing.T) {
	ctrl, _, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/user/register", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	h.RegisterUser(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestRegisterUser_InvalidContentType(t *testing.T) {
	ctrl, _, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "text/plain")

	rr := httptest.NewRecorder()

	h.RegisterUser(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}

func TestRegisterUser_InvalidJSONPayload(t *testing.T) {
	ctrl, _, h := setup(t)
	defer ctrl.Finish()

	req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	h.RegisterUser(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}

func TestRegisterUser_ServiceConflictError(t *testing.T) {
	ctrl, mockService, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	mockService.EXPECT().RegisterUser(gomock.Any(), "user1", "password1").Return(customerror.ErrUniqueKeyConstrantViolation)

	h.RegisterUser(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("Expected status code %v, got %v", http.StatusConflict, status)
	}
}

func TestRegisterUser_ServiceOtherError(t *testing.T) {
	ctrl, mockService, h := setup(t)
	defer ctrl.Finish()

	body := models.User{Login: "user1", Password: "password1"}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("content-type", "application/json")

	rr := httptest.NewRecorder()

	mockService.EXPECT().RegisterUser(gomock.Any(), "user1", "password1").Return(errors.New("service error"))

	h.RegisterUser(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}
