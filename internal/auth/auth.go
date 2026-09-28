// Package auth signs user identifiers and stores them in cookies.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cookieMaxAge = int((30 * 24 * time.Hour) / time.Second)

// CookieName is the name of the authorization cookie.
const CookieName = "gophermart_session"

type userIDKey struct{}

// Manager signs and verifies authorization cookies.
type Manager struct {
	secret []byte
}

// New creates a Manager with a provided secret.
func New(secret []byte) *Manager {
	copied := make([]byte, len(secret))
	copy(copied, secret)
	return &Manager{secret: copied}
}

// NewRandom creates a Manager with a random process-local secret.
func NewRandom() (*Manager, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("read auth secret: %w", err)
	}
	return New(secret), nil
}

// SetCookie writes a signed authorization cookie for userID.
func (m *Manager) SetCookie(w http.ResponseWriter, userID int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    m.sign(userID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   cookieMaxAge,
	})
}

// UserIDFromRequest reads a signed user identifier from request cookies.
func (m *Manager) UserIDFromRequest(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return 0, false
	}

	userID, err := m.verify(cookie.Value)
	return userID, err == nil
}

// WithUserID returns a context containing userID.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserIDFromContext returns userID from ctx.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey{}).(int64)
	return userID, ok
}

func (m *Manager) sign(userID int64) string {
	id := strconv.FormatInt(userID, 10)
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(id))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return id + "." + signature
}

func (m *Manager) verify(value string) (int64, error) {
	id, signature, ok := strings.Cut(value, ".")
	if !ok || id == "" || signature == "" {
		return 0, errors.New("invalid cookie")
	}

	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(id))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return 0, errors.New("invalid signature")
	}

	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid user id")
	}
	return userID, nil
}
