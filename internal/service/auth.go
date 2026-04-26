package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func (s *ShortenerService) Sign(userID string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(userID))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *ShortenerService) Verify(userID string, signature string) bool {
	expected := s.Sign(userID)
	return hmac.Equal([]byte(expected), []byte(signature))
}
