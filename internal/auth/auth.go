package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const CookieName = "session"

// SetAuthCookie выставляет подписанную куку с userID (ID пользователя из БД).
func SetAuthCookie(w http.ResponseWriter, userID int64, secret string) {
	value := cookieValue(userID, secret)
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ValidateCookie проверяет подпись куки и возвращает userID. При невалидной куке — ошибка.
func ValidateCookie(cookie *http.Cookie, secret string) (int64, error) {
	if cookie == nil || cookie.Value == "" {
		return 0, errors.New("empty cookie value")
	}
	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return 0, errors.New("invalid cookie format")
	}
	userIDStr := parts[0]
	signature := parts[1]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid user id in cookie")
	}
	expected := signData(userIDStr, secret)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return 0, errors.New("invalid signature")
	}
	return userID, nil
}

// NewCookie возвращает подписанную куку для userID. Используется в тестах.
func NewCookie(userID int64, secret string) *http.Cookie {
	value := cookieValue(userID, secret)
	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func cookieValue(userID int64, secret string) string {
	s := strconv.FormatInt(userID, 10)
	return s + "." + signData(s, secret)
}

func signData(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
