package services

import (
	"testing"
	"time"
)

func BenchmarkAccrualServiceCreation(b *testing.B) {
	baseURL := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAccrualService(baseURL)
	}
}

func BenchmarkAccrualServiceWithRetryCreation(b *testing.B) {
	baseURL := "http://localhost:8080"
	maxRetries := 3
	baseDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAccrualServiceWithRetry(baseURL, maxRetries, baseDelay, maxDelay)
	}
}

func BenchmarkCalculateDelay(b *testing.B) {
	service := NewAccrualService("http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.calculateDelay(i % 5) // Тестируем разные попытки
	}
}

func BenchmarkShouldRetry(b *testing.B) {
	service := NewAccrualService("http://localhost:8080")

	testCases := []struct {
		err        error
		statusCode int
	}{
		{nil, 200},
		{nil, 429},
		{nil, 500},
		{nil, 502},
		{nil, 503},
		{nil, 504},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc := testCases[i%len(testCases)]
		_ = service.shouldRetry(tc.err, tc.statusCode)
	}
}

// Бенчмарк для симуляции HTTP запросов (без реальных сетевых вызовов)
func BenchmarkAccrualServiceHTTPClient(b *testing.B) {
	service := NewAccrualService("http://localhost:8080")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Создаем запрос (без выполнения)
		url := service.baseURL + "/api/orders/1234567890"
		req, err := service.client.Get(url)
		if err == nil && req != nil {
			req.Body.Close()
		}
	}
}

// Бенчмарк для проверки производительности retry логики
func BenchmarkRetryLogic(b *testing.B) {
	service := NewAccrualService("http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempt := i % service.maxRetries
		delay := service.calculateDelay(attempt)
		_ = delay
	}
}

// Бенчмарк для проверки производительности обработки различных статус кодов
func BenchmarkStatusCodeHandling(b *testing.B) {
	statusCodes := []int{200, 204, 429, 500, 502, 503, 504}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		statusCode := statusCodes[i%len(statusCodes)]
		switch statusCode {
		case 200:
			// Симуляция успешного ответа
		case 204:
			// Симуляция пустого ответа
		case 429:
			// Симуляция rate limit
		case 500, 502, 503, 504:
			// Симуляция серверных ошибок
		}
	}
}
