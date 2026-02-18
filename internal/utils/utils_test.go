package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetOrderInfoKeyPrefix(t *testing.T) {
	prefix := GetOrderInfoKeyPrefix("user1")
	require.Equal(t, "user1_", prefix)
}

func TestGetOrderInfoKey(t *testing.T) {
	key := GetOrderInfoKey("user1", "12345")
	require.Equal(t, "user1_12345", key)
}

func TestValidLuhn_ValidNumber(t *testing.T) {
	// валидный Luhn (классический пример)
	require.True(t, ValidLuhn("79927398713"))
}

func TestValidLuhn_InvalidChecksum(t *testing.T) {
	require.False(t, ValidLuhn("79927398714"))
}

func TestValidLuhn_NonDigit(t *testing.T) {
	require.False(t, ValidLuhn("79927A98713"))
}

func TestValidLuhn_EmptyString(t *testing.T) {
	require.True(t, ValidLuhn(""))
}
