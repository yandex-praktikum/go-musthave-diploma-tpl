package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(password))
	hashPassword := h.Sum(nil)
	hashPasswordString := hex.EncodeToString(hashPassword)
	return hashPasswordString
}
