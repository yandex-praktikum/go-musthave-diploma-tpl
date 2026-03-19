package user

import (
	"context"
	"fmt"
	"gmart/internal/adapters/pgc"
	"gmart/internal/config"
	"gmart/internal/domain"
	"gmart/internal/model/auth"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=user_mock.go  -package=user

// AuthRepoIFace описывает методы репозитория для работы с пользователями и сессиями.
type AuthRepoIFace interface {

	// ExpDuration время протухания сессионной куки
	ExpDuration() time.Duration

	// SignUp регистрирует нового пользователя и возвращает его внутренний ID.
	// Возвращает ErrUserAlreadyExists, если логин занят.
	SignUp(ctx context.Context, login domain.Login, password domain.Password, ua domain.UserAgent, ip string) (domain.UserID, domain.SessionID, error)

	// SignIn проверяет пароль (bcrypt) и создает запись в таблице сессий.
	// Возвращает SessionID (UUID), UserID и ошибку (ErrUserNotFound или ErrInvalidPassword).
	SignIn(ctx context.Context, login domain.Login, password domain.Password, userAgent domain.UserAgent, ip string) (domain.UserID, domain.SessionID, error)

	// RefreshSession продлевает время жизни существующей сессии.
	RefreshSession(ctx context.Context, sessionID domain.SessionID, userAgent domain.UserAgent, ip string) (domain.UserID, error)

	// SignOut удаляет сессию из базы данных.
	SignOut(ctx context.Context, sessionID domain.SessionID) error

	// SignInSleep — обертка над SignIn с искусственной задержкой для защиты от Timing Attacks.
	// Используется для выравнивания времени ответа при отсутствии пользователя.
	SignInSleep(ctx context.Context, login domain.Login, password domain.Password, ua domain.UserAgent, ip string) (domain.UserID, domain.SessionID, error)
}

// TokenGeneratorIFace описывает контракт для создания JWT токенов.
// Позволяет подменить реальную логику подписи в юнит-тестах.
type TokenGeneratorIFace interface {
	// Generate создает подписанный токен для конкретного пользователя.
	Generate(userID domain.UserID) (domain.Token, error)
}

// ============ Class ============

// UserLogin класс манипулирования пользователями
type User struct {
	authRepo       AuthRepoIFace
	tokenGenerator TokenGeneratorIFace
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

// SignUp регистрирует пользователя
func (u *User) SignUp(ctx context.Context, in *signUpInput) (*signOutput, error) {

	// Регистрируем пользователя в БД и проверяем ошибки
	userID, sessionID, err := u.authRepo.SignUp(ctx, in.Body.Login, in.Body.Password, in.UserAgent, in.XForwardedFor)
	if err != nil {
		return nil, fmt.Errorf("sign up fail: %w", err)
	}

	// Генерируем токены и куки
	return u.makeAuthResponse(userID, sessionID)
}

// SignIn залогин пользователя
func (u *User) SignIn(ctx context.Context, in *signInInput) (*signOutput, error) {

	var (
		userID    domain.UserID
		sessionID domain.SessionID
		err       error
	)

	// Проверяем наличие сессионной куки (Fast Track)
	if in.SessionID.UUID() == uuid.Nil {
		userID, sessionID, err = u.authRepo.SignInSleep(ctx, in.Body.Login, in.Body.Password, in.UserAgent, in.XForwardedFor)
	} else {
		if userID, err = u.authRepo.RefreshSession(ctx, in.SessionID, in.UserAgent, in.XForwardedFor); err != nil {
			userID, sessionID, err = u.authRepo.SignInSleep(ctx, in.Body.Login, in.Body.Password, in.UserAgent, in.XForwardedFor)
		} else {
			sessionID = in.SessionID
		}
	}
	if err != nil {
		return nil, fmt.Errorf("sign in fail: %w", err)
	}

	// Генерируем токены и куки
	return u.makeAuthResponse(userID, sessionID)
}

// makeAuthResponse — единая точка формирования авторизационных данных
func (u *User) makeAuthResponse(userID domain.UserID, sessionID domain.SessionID) (*signOutput, error) {
	jwtToken, err := u.tokenGenerator.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("JWT token generate fail: %w", err)
	}

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(u.authRepo.ExpDuration().Seconds()),
	}

	res := &signOutput{
		Authorization: "Bearer " + jwtToken.String(),
		SetCookie:     cookie.String(),
	}
	res.Body.Code = http.StatusOK
	res.Body.Status = "Success"
	res.Body.Token = jwtToken

	return res, nil
}
