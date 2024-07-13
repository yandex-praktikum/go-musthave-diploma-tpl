package auth

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	privateRouter "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/middlware/private_router"
	"log"
	"net/http"
	"time"
)

func SetAuthCookie(w http.ResponseWriter, authToken string, tokenExp time.Duration) {
	//tokenExp = tokenExp
	cookie := http.Cookie{
		Name:  config.AuthCookie, // accessToken
		Value: authToken,
		Path:  "/",
		//MaxAge:   int(tokenExp),
		//HttpOnly: true,
		//Secure:   true,
		//SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}

func GetUserIDFromContext(ctx context.Context) *int {
	userIDAny, ok := ctx.Value(privateRouter.KeyUserID).(int)
	if ok {
		return &userIDAny
	}

	log.Fatalln("нет userID")
	return nil
}
