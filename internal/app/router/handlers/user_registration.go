package handlers

import (
	"encoding/json"
	"errors"
	resp "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/router/responses"
	s "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/service"
	m "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/service/methods"
	"io"
	"net/http"
)

// UserRegister регистрация пользователя
/*
POST /api/user/register HTTP/1.1
Content-Type: application/json
...

{
	"login": "<login>",
	"password": "<password>"
}

Возможные коды ответа:

200 — пользователь успешно зарегистрирован и аутентифицирован;
400 — неверный формат запроса;
409 — логин уже занят;
500 — внутренняя ошибка сервера.
*/

func UserRegister(w http.ResponseWriter, r *http.Request) {
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
		resp.NoContent(w, http.StatusBadRequest)
		return
	}

	encrP, err := s.EncryptPass(b.Password)
	if err != nil {
		resp.NoContent(w, http.StatusInternalServerError)
		return
	}
	u := m.NewUser(b.Login, encrP)

	if err = u.Add(); err != nil {
		if errors.Is(err, m.ErrLoginUnavailable) {
			resp.NoContent(w, http.StatusConflict)
			return
		}
		resp.NoContent(w, http.StatusInternalServerError)
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
