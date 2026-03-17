package user

import (
	"context"
	"fmt"
	"gmart/internal/adapters/pgc"
	"gmart/internal/config"
	"gmart/internal/domain"
	"gmart/internal/model/auth"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// ============ Class ============

// UserLogin класс манипулирования пользователями
type User struct {
	authRepo       *AuthRepo
	tokenGenerator *auth.TokenGenerator
}

// NewUser создает новый объект манипулирования пользователями
func NewUser(cfg config.Config, pg pgc.PgInstance, m AuthMetrics) (*User, error) {

	// Токен генератор
	tGen, err := auth.NewTokenGenerator(jwt.SigningMethodHS256, cfg.JWTSecretKey(), cfg.JWTTTL())
	if err != nil {
		return nil, fmt.Errorf("token generator create fail: %w", err)
	}

	// Результат
	return &User{
		authRepo:       NewAuthRepo(pg, cfg.SessTTL(), m),
		tokenGenerator: tGen,
	}, nil

}

// ============ UseCase ============

// SugnUp регистрирует пользователя
func (u *User) SignUp(ctx context.Context, in *signUpInput) (*signUpOutput, error) {

	// Регистрируем пользователя в БД и проверяем ошибки
	_, err := u.authRepo.SignUp(ctx, domain.Login(in.Body.Login), domain.Password(in.Body.Password))
	if err != nil {
		return nil, fmt.Errorf("sign up fail: %w", err)
	}

	// Логинимся по логин/паролю
	sessionID, userID, err := u.authRepo.SignIn(ctx, domain.Login(in.Body.Login), domain.Password(in.Body.Password), domain.UserAgent(in.UserAgent), in.XForwardedFor)
	if err != nil {
		return nil, fmt.Errorf("sign in fail: %w", err)
	}

	// Формируем короткоживущий JWT
	jwtToken, err := u.tokenGenerator.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("JWT token generate fail: %w", err)
	}

	// Формируем долгоживущую сессонную куку
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	res := &signUpOutput{
		Authorization: "Bearer " + jwtToken.String(),
		SetCookie:     cookie.String(),
	}
	res.Body.Token = jwtToken.String()

	return res, nil
}

// SugnIn залогин пользователя
func (u *User) SignIn(ctx context.Context, in *signInInput) (*signInOutput, error) {

	// Логинимся по логин/паролю, используя устойчивую ко взлому проверку пароля
	sessionID, userID, err := u.authRepo.SignInSleep(ctx, domain.Login(in.Body.Login), domain.Password(in.Body.Password), domain.UserAgent(in.UserAgent), in.XForwardedFor)
	if err != nil {
		return nil, fmt.Errorf("sign in fail: %w", err)
	}

	// Короткоживущий JWT
	jwtToken, err := u.tokenGenerator.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("JWT token generate fail: %w", err)
	}

	// Долгоживущая кука
	// Формируем долгоживущую сессонную куку
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	// Ответ
	res := &signInOutput{
		Authorization: "Bearer " + jwtToken.String(),
		SetCookie:     cookie.String(),
	}
	res.Body.Token = jwtToken.String()

	return res, nil
}
