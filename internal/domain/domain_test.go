package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestUserPasswordHiddenInJSON(t *testing.T) {
	u := User{ID: 1, Login: "alice", Password: "hash", CreatedAt: time.Now()}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if _, ok := m["password"]; ok {
		t.Error("password field should not be present in JSON")
	}
	if m["login"] != "alice" {
		t.Errorf("login = %v, want alice", m["login"])
	}
}

func TestOrderResponseAccrualOmitted(t *testing.T) {
	accrual := 100.5
	withAccrual, _ := json.Marshal(OrderResponse{Number: "1", Status: "PROCESSED", Accrual: &accrual})
	withoutAccrual, _ := json.Marshal(OrderResponse{Number: "1", Status: "NEW"})

	if len(withAccrual) <= len(withoutAccrual) {
		t.Error("accrual field should be present when set")
	}
	if jsonContains(withoutAccrual, "accrual") {
		t.Error("accrual field should be omitted when nil")
	}
}

func TestWithdrawRequestJSON(t *testing.T) {
	data := []byte(`{"order":"123","sum":50.5}`)
	var req WithdrawRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if req.Order != "123" || req.Sum != 50.5 {
		t.Errorf("unexpected request: %+v", req)
	}
}

func TestAccrualStatusConstants(t *testing.T) {
	accrualStatuses := map[string]string{
		AccrualStatusRegistered: "REGISTERED",
		AccrualStatusInvalid:    "INVALID",
		AccrualStatusProcessing: "PROCESSING",
		AccrualStatusProcessed:  "PROCESSED",
	}
	for got, want := range accrualStatuses {
		if got != want {
			t.Errorf("constant = %q, want %q", got, want)
		}
	}

	orderStatuses := map[string]string{
		OrderStatusNew:        "NEW",
		OrderStatusProcessing: "PROCESSING",
		OrderStatusInvalid:    "INVALID",
		OrderStatusProcessed:  "PROCESSED",
	}
	for got, want := range orderStatuses {
		if got != want {
			t.Errorf("constant = %q, want %q", got, want)
		}
	}
}

func TestBusinessErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrLoginAlreadyExists,
		ErrInvalidCredentials,
		ErrInvalidOrderNumber,
		ErrOrderAlreadyUploadedByUser,
		ErrOrderAlreadyUploadedByAnotherUser,
		ErrOrderNotFound,
		ErrInsufficientBalance,
		ErrUserNotFound,
		ErrUnauthorized,
		ErrAccrualServiceUnavailable,
	}
	seen := make(map[string]bool, len(errs))
	for _, err := range errs {
		if err == nil {
			t.Fatal("business error must not be nil")
		}
		msg := err.Error()
		if seen[msg] {
			t.Errorf("duplicate error message: %q", msg)
		}
		seen[msg] = true
	}
	if errors.Is(ErrOrderNotFound, ErrUserNotFound) {
		t.Error("distinct errors should not match each other")
	}
}

// jsonContains проверяет наличие подстроки в JSON-представлении.
func jsonContains(data []byte, substr string) bool {
	return len(data) > 0 && string(data) != "" && contains(string(data), substr)
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
