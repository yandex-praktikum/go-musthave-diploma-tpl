package auth

import (
	"time"

	"github.com/RedWood011/cmd/gophermart/internal/config"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func CreateTokenJWT(user entity.User, cfg config.TokenConfig) (string, error) {
	exp := time.Now().Add(cfg.AccessTimeLiveToken).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["exp"] = exp
	//fmt.Println(user.ID)
	t, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}

func MiddlewareJWT(cfg config.TokenConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			return []byte(cfg.SecretKey), nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalide Token",
			})
		}
		claims := token.Claims.(jwt.MapClaims)
		userID := claims[cfg.UserKey].(string)
		c.Locals(cfg.UserKey, userID)
		return c.Next()
	}
}
