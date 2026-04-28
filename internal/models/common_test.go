package models_test

import (
	"testing"

	"github.com/KalessinD/gophermart/internal/models"
)

// TestFromJSON_Invalid проверяет обработку невалидного JSON
func TestFromJSON_Invalid(t *testing.T) {
	invalidJSON := `{"id": "123", "login": ` // Оборванный JSON

	_, err := models.FromJSON[models.User]([]byte(invalidJSON))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON, but got nil")
	}
}
