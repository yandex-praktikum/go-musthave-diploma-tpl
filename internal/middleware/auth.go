package middleware

import (
	"errors"
	"net/http"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/auth"
	"github.com/gin-gonic/gin"
)

// KeyUserID — ключ в gin.Context, под которым сохраняется userID (int64) после успешной проверки куки.
const KeyUserID = "user_id"

// GetUserID достаёт userID из контекста (должен быть установлен RequireAuth). Для защищённых ручек.
func GetUserID(c *gin.Context) (int64, error) {
	v, ok := c.Get(KeyUserID)
	if !ok {
		return 0, errors.New("user_id not in context")
	}
	userID, ok := v.(int64)
	if !ok {
		return 0, errors.New("user_id invalid type")
	}
	return userID, nil
}

// RequireAuth возвращает Gin middleware: проверяет куку, при валидной куке кладёт userID в контекст и вызывает Next.
// При отсутствии или невалидной куке отвечает 401 и прерывает цепочку.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(auth.CookieName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID, err := auth.ValidateCookie(cookie, secret)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(KeyUserID, userID)
		c.Next()
	}
}
