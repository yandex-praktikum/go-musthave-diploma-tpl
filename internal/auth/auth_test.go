package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieRoundTrip(t *testing.T) {
	manager := New([]byte("test-secret"))

	response := httptest.NewRecorder()
	manager.SetCookie(response, 42)

	request := httptest.NewRequest("GET", "/", nil)
	request.AddCookie(response.Result().Cookies()[0])

	userID, ok := manager.UserIDFromRequest(request)
	if !ok || userID != 42 {
		t.Fatalf("userID = %d, ok = %v", userID, ok)
	}
}

func TestCookieTampering(t *testing.T) {
	manager := New([]byte("test-secret"))

	request := httptest.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{Name: CookieName, Value: "42.bad-signature"})

	if _, ok := manager.UserIDFromRequest(request); ok {
		t.Fatal("tampered cookie was accepted")
	}
}
