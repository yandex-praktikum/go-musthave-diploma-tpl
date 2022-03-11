package cache

import (
	"fmt"
	"sync"
)

type Cache struct {
	cash *sync.Map
}

func NewCache() *Cache {
	return &Cache{
		cash: &sync.Map{},
	}
}

//add to cash ============================================================
func (c *Cache) AddToCache(key string, value string) {
	c.cash.Store(key, value)
}

//get from cash ==========================================================
func (c *Cache) GetFromCache(key string) (string, bool) {
	value, ok := c.cash.Load(key)
	return fmt.Sprint(value), ok
}

//remove from cash =======================================================
func (r *Cache) RemoveFromCache(key string) {
	r.cash.Delete(key)
}

//print cash =============================================================
func (c *Cache) PrintCache() string {
	return fmt.Sprintf("cash: %v", c.cash)
}
