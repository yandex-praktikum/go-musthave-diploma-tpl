package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	mw "github.com/KalessinD/gophermart/internal/middleware"
	"github.com/go-chi/chi/middleware"
)

func TestMiddleware_LogsRequest(t *testing.T) {
	// создаём observer для перехвата логов
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	// тестовый handler
	nextHandlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// оборачиваем middleware
	middleware := mw.Logger(logger)
	handler := middleware(next)

	// создаём тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// вызываем
	handler.ServeHTTP(rec, req)

	if !nextHandlerCalled {
		t.Fatal("next handler was not called")
	}

	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}

	entry := logs[0]

	if entry.Message != "request completed" {
		t.Fatalf("unexpected log message: %s", entry.Message)
	}

	fields := entry.ContextMap()

	if fields["method"] != http.MethodGet {
		t.Errorf("expected method %s, got %v", http.MethodGet, fields["method"])
	}

	if fields["path"] != "/test" {
		t.Errorf("expected path /test, got %v", fields["path"])
	}

	if fields["remote_addr"] == "" {
		t.Error("expected remote_addr to be set")
	}

	if fields["duration"] == nil {
		t.Error("expected duration field")
	}
}

// TestGetLogger тестирует получение логгера из контекста
func TestGetLogger(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Logger exists in context", func(t *testing.T) {
		ctx := mw.AddLoggerToContext(t.Context(), logger)
		got := mw.GetLogger(ctx)
		if got == nil {
			t.Error("expected logger, got nil")
		}
	})

	t.Run("Logger not in context", func(t *testing.T) {
		got := mw.GetLogger(t.Context())
		if got != nil {
			t.Error("expected nil, got logger")
		}
	})
}

// TestGetEncodingField тестирует приватную функцию getEncodingField
func TestGetEncodingField(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		headerName string
		headerVal  []string
		wantSkip   bool
	}{
		{
			name:       "Existing header",
			prefix:     "response-",
			headerName: "Content-Encoding",
			headerVal:  []string{"gzip"},
			wantSkip:   false,
		},
		{
			name:       "Missing header",
			prefix:     "response-",
			headerName: "Accept-Encoding",
			headerVal:  nil,
			wantSkip:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := make(http.Header)
			if tt.headerVal != nil {
				for _, v := range tt.headerVal {
					h.Add(tt.headerName, v)
				}
			}

			field := mw.GetEncodingField(tt.prefix, tt.headerName, h)

			if tt.wantSkip {
				if field.Type != zap.Skip().Type {
					t.Errorf("expected Skip field, got %v", field)
				}
			} else {
				if field.Type == zap.Skip().Type {
					t.Errorf("expected String field, got Skip")
				}
			}
		})
	}
}

// TestMiddleware тестирует полную работу мидлвари
func TestMiddleware(t *testing.T) {
	// Используем observer, чтобы перехватить логи
	observerCore, logs := observer.New(zap.InfoLevel)
	testLogger := zap.New(observerCore)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что логгер попал в контекст внутри хендлера
		log := mw.GetLogger(r.Context())
		if log == nil {
			t.Error("logger not found in request context inside handler")
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	// Добавляем RequestID в контекст запроса
	reqID := "test-request-id-123"
	ctx := context.WithValue(t.Context(), middleware.RequestIDKey, reqID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Запускаем мидлварю
	mw := mw.Logger(testLogger)
	mw(handler).ServeHTTP(rec, req)

	// Проверяем, что ответ прошел корректно
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	// Проверяем логи
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Message != "request completed" {
		t.Errorf("unexpected message: %s", entry.Message)
	}

	// Вспомогательная функция для поиска поля в логе
	findField := func(key string) (zap.Field, bool) {
		for _, f := range entry.Context {
			if f.Key == key {
				return f, true
			}
		}
		return zap.Field{}, false
	}

	// Проверяем наличие и значение ключевых полей
	tests := []struct {
		key      string
		expected any
	}{
		{"request_id", reqID},
		{"method", http.MethodPost},
		{"path", "/test"},
		{"status", int64(http.StatusCreated)},
		{"response_size", int64(len("created"))},
	}

	for _, tt := range tests {
		t.Run("check field "+tt.key, func(t *testing.T) {
			field, ok := findField(tt.key)
			if !ok {
				t.Errorf("field %s not found in log", tt.key)
				return
			}

			// Сравниваем значения в зависимости от типа
			switch expectedValue := tt.expected.(type) {
			case string:
				if field.String != expectedValue {
					t.Errorf("field %s: expected %s, got %s", tt.key, expectedValue, field.String)
				}
			case int64:
				if field.Integer != expectedValue {
					t.Errorf("field %s: expected %d, got %d", tt.key, expectedValue, field.Integer)
				}
			}
		})
	}
}
