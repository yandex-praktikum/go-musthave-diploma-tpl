package auth

import (
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"net/http"
	"time"
)

func SetAuthCookie(w http.ResponseWriter, authToken string, tokenExp time.Duration) {
	//tokenExp = tokenExp
	cookie := http.Cookie{
		Name:  config.AUTH_COOKIE, // accessToken
		Value: authToken,
		Path:  "/",
		//MaxAge:   int(tokenExp),
		//HttpOnly: true,
		//Secure:   true,
		//SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
