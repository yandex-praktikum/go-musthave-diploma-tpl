package auth

import (
	"net/http"
	"time"
)

func SetAuthCookie(w http.ResponseWriter, authToken string, tokenExp time.Duration) {
	cookie := http.Cookie{
		Name:  "Auth", // accessToken
		Value: authToken,
		Path:  "/",
		//MaxAge:   int(tokenExp),
		//HttpOnly: true,
		//Secure:   true,
		//SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
