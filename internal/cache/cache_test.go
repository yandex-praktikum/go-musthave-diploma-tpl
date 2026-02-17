package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache[string]()

	cache.Set("k1", "value")

	val, ok := cache.Get("k1")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

func TestCache_Get_NotFound(t *testing.T) {
	cache := NewCache[string]()

	val, ok := cache.Get("unknown")

	assert.False(t, ok)
	assert.Equal(t, "", val) // zero value для string
}

func TestCache_GetByPrefix(t *testing.T) {
	cache := NewCache[int]()

	cache.Set("u1_1", 10)
	cache.Set("u1_2", 20)
	cache.Set("u2_1", 30)

	res := cache.GetByPrefix("u1_")

	assert.Len(t, res, 2)
	assert.Contains(t, res, 10)
	assert.Contains(t, res, 20)
}

func TestCache_GetByPrefix_Empty(t *testing.T) {
	cache := NewCache[int]()

	cache.Set("u1_1", 10)

	res := cache.GetByPrefix("u2_")

	assert.Len(t, res, 0)
}
