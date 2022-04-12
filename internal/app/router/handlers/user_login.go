package handlers

import (
	"encoding/json"
	"errors"
	resp "github.com/EestiChameleon/GOphermart/internal/app/router/responses"
	s "github.com/EestiChameleon/GOphermart/internal/app/service"
	"github.com/EestiChameleon/GOphermart/internal/app/service/methods"
	db "github.com/EestiChameleon/GOphermart/internal/app/storage"
	"io"
	"net/http"
)

// UserLogin аутентификация пользователя;
/*
POST /api/user/login HTTP/1.1
Content-Type: application/json
...

{
	"login": "<login>",
	"password": "<password>"
}
Возможные коды ответа:

200 — пользователь успешно аутентифицирован;
400 — неверный формат запроса;
401 — неверная пара логин/пароль;
500 — внутренняя ошибка сервера.
*/

func UserLogin(w http.ResponseWriter, r *http.Request) {
	var b resp.LoginData
	data, err := io.ReadAll(r.Body)
	if err != nil {
		resp.NoContent(w, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, &b)
	if err != nil {
		resp.NoContent(w, http.StatusBadRequest)
		return
	}

	if b.Password == "" || b.Login == "" {
		resp.NoContent(w, http.StatusBadRequest) // 401?
		return
	}

	u := methods.NewUser(b.Login, b.Password)
	if err = u.GetByLogin(); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			resp.NoContent(w, http.StatusUnauthorized)
			return
		}
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}

	encrP, err := s.EncryptPass(b.Password)
	if err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}

	if encrP != u.Password {
		resp.NoContent(w, http.StatusUnauthorized)
		return
	}

	token, err := s.JWTEncode("userID", u.ID)
	if err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, resp.CreateCookie("gophermartID", token))
	resp.NoContent(w, http.StatusOK)
}
