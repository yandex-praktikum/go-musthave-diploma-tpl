package cache

import (
	"fmt"
	"strings"
	"sync"
)

type Cache[T any] struct {
	mu   sync.RWMutex
	data map[string]T
}

func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		data: make(map[string]T),
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, founded := c.data[key]
	return value, founded
}

func (c *Cache[T]) GetByPrefix(prefix string) []T {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := []T{}
	for key, order := range c.data {
		if strings.HasPrefix(key, prefix) {
			data = append(data, order)
		}
	}

	return data
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Cache[T]) UpdValue(key string, action func(T)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.data[key]
	fmt.Println(key)
	if !ok {
		fmt.Println("not found")
		return
	}
	action(value)
}
