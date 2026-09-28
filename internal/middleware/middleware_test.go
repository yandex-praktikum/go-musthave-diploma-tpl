package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
)

func TestGzipRequest(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, _ = writer.Write([]byte("12345678903"))
	_ = writer.Close()

	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "12345678903" {
			t.Fatalf("body = %q", body)
		}
		w.WriteHeader(http.StatusAccepted)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/user/orders", &compressed)
	request.Header.Set("Content-Encoding", "gzip")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestGzipResponse(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q", response.Header().Get("Content-Encoding"))
	}
	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("body = %q", body)
	}
}

func TestGzipNoContentResponseStaysEmpty(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if response.Header().Get("Content-Encoding") != "" {
		t.Fatalf("Content-Encoding = %q", response.Header().Get("Content-Encoding"))
	}
	if response.Body.Len() != 0 {
		t.Fatalf("body is not empty: %q", response.Body.String())
	}
}

func TestGzipBrokenRequest(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("broken"))
	request.Header.Set("Content-Encoding", "gzip")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestLoggingDoesNotBreakResponse(t *testing.T) {
	handler := Logging(zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("ok"))
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusTeapot || response.Body.String() != "ok" {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

func TestAuthProtectsRoute(t *testing.T) {
	manager := auth.New([]byte("test-secret"))
	handler := Auth(manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok || userID != 7 {
			t.Fatalf("userID = %d, ok = %v", userID, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status without cookie = %d", response.Code)
	}

	cookieResponse := httptest.NewRecorder()
	manager.SetCookie(cookieResponse, 7)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookieResponse.Result().Cookies()[0])

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status with cookie = %d", response.Code)
	}
}
