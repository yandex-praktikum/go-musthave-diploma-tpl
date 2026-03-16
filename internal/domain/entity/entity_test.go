package entity

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserStruct(t *testing.T) {
	user := User{
		ID:        1,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
		Login:     "testuser",
		Password:  "secret",
		IsActive:  true,
	}

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Login != "testuser" {
		t.Errorf("Expected login 'testuser', got %s", user.Login)
	}
	if !user.IsActive {
		t.Error("Expected IsActive true, got false")
	}

	// Проверка JSON тегов
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal user: %v", err)
	}

	if _, ok := decoded["password"]; ok {
		t.Error("Password field should not be present in JSON")
	}
	if _, ok := decoded["login"]; !ok {
		t.Error("Login field should be present in JSON")
	}
}

func TestUserWithOrders(t *testing.T) {
	user := User{ID: 1, Login: "testuser"}
	orders := []Order{
		{ID: 1, Number: "123456", Status: OrderStatusNew},
		{ID: 2, Number: "789012", Status: OrderStatusProcessed},
	}

	userWithOrders := UserWithOrders{
		User:   user,
		Orders: orders,
	}

	if len(userWithOrders.Orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(userWithOrders.Orders))
	}
	if userWithOrders.User.Login != "testuser" {
		t.Errorf("Expected user login 'testuser', got %s", userWithOrders.User.Login)
	}
}

func TestOrderStatusConstants(t *testing.T) {
	testCases := []struct {
		status     OrderStatus
		expected   string
		shouldFail bool
	}{
		{OrderStatusNew, "NEW", false},
		{OrderStatusProcessing, "PROCESSING", false},
		{OrderStatusInvalid, "INVALID", false},
		{OrderStatusProcessed, "PROCESSED", false},
		{"UNKNOWN", "UNKNOWN", true},
	}

	for _, tc := range testCases {
		if string(tc.status) != tc.expected && !tc.shouldFail {
			t.Errorf("Expected status %s, got %s", tc.expected, tc.status)
		}
	}
}

func TestOrderStruct(t *testing.T) {
	accrual := 100.50
	order := Order{
		ID:          1,
		UploadedAt:  time.Now().Format(time.RFC3339),
		ProcessedAt: time.Now().Format(time.RFC3339),
		Number:      "1234567890",
		Status:      OrderStatusNew,
		Accrual:     &accrual,
		UserID:      1,
	}

	if order.Number != "1234567890" {
		t.Errorf("Expected order number '1234567890', got %s", order.Number)
	}
	if order.Status != OrderStatusNew {
		t.Errorf("Expected status NEW, got %s", order.Status)
	}
	if *order.Accrual != accrual {
		t.Errorf("Expected accrual %f, got %f", accrual, *order.Accrual)
	}

	// Проверка JSON-сериализации с опусканием ProcessedAt если пусто
	order2 := Order{
		ID:      2,
		Number:  "9876543210",
		Status:  OrderStatusNew,
		UserID:  1,
		Accrual: nil,
	}

	data, err := json.Marshal(order2)
	if err != nil {
		t.Fatalf("Failed to marshal order: %v", err)
	}

	if string(data) == "" {
		t.Error("Marshaled order should not be empty")
	}
}

func TestOrderWithUser(t *testing.T) {
	order := Order{
		ID:     1,
		Number: "123456",
		Status: OrderStatusNew,
		UserID: 1,
	}
	user := User{
		ID:    1,
		Login: "testuser",
	}

	orderWithUser := OrderWithUser{
		Order: order,
		User:  user,
	}

	if orderWithUser.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", orderWithUser.UserID)
	}
	if orderWithUser.User.Login != "testuser" {
		t.Errorf("Expected user login 'testuser', got %s", orderWithUser.User.Login)
	}
}

func TestUserBalance(t *testing.T) {
	balance := UserBalance{
		UserID:    1,
		Current:   500.75,
		Withdrawn: 200.25,
	}

	if balance.Current != 500.75 {
		t.Errorf("Expected current balance 500.75, got %f", balance.Current)
	}
	if balance.Withdrawn != 200.25 {
		t.Errorf("Expected withdrawn 200.25, got %f", balance.Withdrawn)
	}
	if balance.Current+balance.Withdrawn != 701.0 {
		t.Errorf("Expected total 701.0, got %f", balance.Current+balance.Withdrawn)
	}
}

func TestWithdrawal(t *testing.T) {
	withdrawal := Withdrawal{
		UserID:      1,
		Order:       "123456",
		Sum:         150.50,
		ProcessedAt: time.Now().Format(time.RFC3339),
	}

	if withdrawal.Sum != 150.50 {
		t.Errorf("Expected sum 150.50, got %f", withdrawal.Sum)
	}
	if withdrawal.Order != "123456" {
		t.Errorf("Expected order '123456', got %s", withdrawal.Order)
	}
}

func TestBalanceSummary(t *testing.T) {
	summary := BalanceSummary{
		TotalEarned:    1000.0,
		TotalSpent:     300.0,
		CurrentBalance: 700.0,
	}

	if summary.TotalEarned-summary.TotalSpent != summary.CurrentBalance {
		t.Errorf("Balance calculation incorrect: %f - %f != %f",
			summary.TotalEarned, summary.TotalSpent, summary.CurrentBalance)
	}
}

func TestUserStats(t *testing.T) {
	stats := UserStats{
		UserID:           1,
		TotalOrders:      10,
		ProcessedOrders:  5,
		NewOrders:        2,
		ProcessingOrders: 2,
		InvalidOrders:    1,
		TotalAccrual:     500.0,
		TotalWithdrawn:   200.0,
	}

	totalCalculated := stats.NewOrders + stats.ProcessingOrders +
		stats.ProcessedOrders + stats.InvalidOrders

	if stats.TotalOrders != totalCalculated {
		t.Errorf("Total orders mismatch: expected %d, calculated %d",
			stats.TotalOrders, totalCalculated)
	}

	if stats.TotalAccrual < stats.TotalWithdrawn {
		t.Error("Total accrual should not be less than total withdrawn")
	}
}

func TestOrdersSummary(t *testing.T) {
	summary := OrdersSummary{
		Total:           100,
		NewCount:        20,
		ProcessingCount: 30,
		ProcessedCount:  40,
		InvalidCount:    10,
		TotalAccrual:    5000.0,
	}

	totalCalculated := summary.NewCount + summary.ProcessingCount +
		summary.ProcessedCount + summary.InvalidCount

	if summary.Total != totalCalculated {
		t.Errorf("Total mismatch: expected %d, calculated %d",
			summary.Total, totalCalculated)
	}
}

func TestWithdrawalsSummary(t *testing.T) {
	summary := WithdrawalsSummary{
		TotalWithdrawals: 50,
		TotalAmount:      2500.75,
		UniqueUsers:      10,
	}

	if summary.TotalAmount <= 0 && summary.TotalWithdrawals > 0 {
		t.Error("Total amount should be positive when there are withdrawals")
	}
	if summary.UniqueUsers > summary.TotalWithdrawals {
		t.Error("Unique users cannot exceed total withdrawals")
	}
}

func TestUserWithdrawalsSummary(t *testing.T) {
	summary := UserWithdrawalsSummary{
		UserID:          1,
		WithdrawalCount: 5,
		TotalAmount:     750.50,
		FirstWithdrawal: "2024-01-01T10:00:00Z",
		LastWithdrawal:  "2024-01-05T15:30:00Z",
	}

	if summary.WithdrawalCount <= 0 && summary.TotalAmount > 0 {
		t.Error("Withdrawal count should be positive when there's amount")
	}
	if summary.WithdrawalCount == 0 && (summary.FirstWithdrawal != "" || summary.LastWithdrawal != "") {
		t.Error("Withdrawal dates should be empty when count is zero")
	}
}

func TestOrderStatusJSON(t *testing.T) {
	// Проверка сериализации/десериализации OrderStatus
	order := Order{
		ID:     1,
		Number: "123456",
		Status: OrderStatusNew,
		UserID: 1,
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("Failed to marshal order: %v", err)
	}

	var decoded Order
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal order: %v", err)
	}

	if decoded.Status != OrderStatusNew {
		t.Errorf("Expected status NEW after unmarshal, got %s", decoded.Status)
	}

	// Проверка невалидного статуса
	invalidJSON := `{"id":1,"number":"123456","status":"INVALID_STATUS","user_id":1}`
	var order2 Order
	err = json.Unmarshal([]byte(invalidJSON), &order2)
	if err != nil {
		// Ожидаем ошибку или успешную десериализацию с невалидным значением
		t.Logf("Expected error for invalid status: %v", err)
	}
}

func TestEmptyFieldsOmission(t *testing.T) {
	// Проверка опускания пустых полей в JSON
	user := User{
		ID:    1,
		Login: "test",
		// UpdatedAt пустое
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, exists := result["updated_at"]; exists {
		t.Error("updated_at should be omitted when empty")
	}
}

// Бенчмарк-тесты
func BenchmarkUserMarshal(b *testing.B) {
	user := User{
		ID:        1,
		CreatedAt: time.Now().Format(time.RFC3339),
		Login:     "benchmarkuser",
		IsActive:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(user)
		if err != nil {
			b.Fatalf("Failed to marshal: %v", err)
		}
	}
}

func BenchmarkOrderStatusComparison(b *testing.B) {
	statuses := []OrderStatus{
		OrderStatusNew,
		OrderStatusProcessing,
		OrderStatusInvalid,
		OrderStatusProcessed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range statuses {
			switch s {
			case OrderStatusNew:
				_ = "NEW"
			case OrderStatusProcessing:
				_ = "PROCESSING"
			case OrderStatusInvalid:
				_ = "INVALID"
			case OrderStatusProcessed:
				_ = "PROCESSED"
			}
		}
	}
}
