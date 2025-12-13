package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// ComputeHMACSHA256 вычисляет HMAC-SHA256 хеш для данных с использованием ключа
func ComputeHMACSHA256(data []byte, key string) string {
	if key == "" {
		return "" // Если ключ не установлен, возвращаем пустую строку
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// VerifyHMACSHA256 проверяет, соответствует ли хеш данным и ключу
func VerifyHMACSHA256(data []byte, receivedHash string, key string) bool {
	if key == "" || receivedHash == "" {
		return true // Если ключ или хеш не установлены, пропускаем проверку
	}

	expectedHash := ComputeHMACSHA256(data, key)
	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}
