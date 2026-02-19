package cookie

import (
	"time"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/jellydator/ttlcache/v3"
)

// Хэндле для работы с кукесами, они хранятся в течении 24 часов
type CookieHandler struct {
	cache *ttlcache.Cache[string, *dto.UserData]
}

func NewCookieHandler() *CookieHandler {
	cache := ttlcache.New[string, *dto.UserData]()

	return &CookieHandler{
		cache: cache,
	}
}

func (c *CookieHandler) Set(sessionId string, userData *dto.UserData) {
	c.cache.Set(sessionId, userData, 24*time.Hour)
}

func (c *CookieHandler) Get(sessionId string) (*dto.UserData, bool) {
	cookie := c.cache.Get(sessionId)
	if cookie == nil {
		return nil, false
	}
	return cookie.Value(), true
}
