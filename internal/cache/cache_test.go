package cache

import (
	"testing"

	"github.com/Raime-34/gophermart.git/internal/dto"
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

func TestCache_UpdValue_ExistingKey(t *testing.T) {
	c := &Cache[*dto.OrderInfo]{data: make(map[string]*dto.OrderInfo)}
	key := "k1"
	c.data[key] = &dto.OrderInfo{Status: "OLD"}

	c.UpdValue(key, func(v *dto.OrderInfo) {
		v.Status = "NEW"
	})

	assert.Equal(t, "NEW", c.data[key].Status)
}

func TestCache_UpdValue_MissingKey(t *testing.T) {
	c := &Cache[*dto.OrderInfo]{data: make(map[string]*dto.OrderInfo)}
	key := "missing"

	c.UpdValue(key, func(v *dto.OrderInfo) {
		t.Fatal("action must not be called")
	})
}
