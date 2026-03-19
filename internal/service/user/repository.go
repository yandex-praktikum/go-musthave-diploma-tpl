package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"gmart/internal/adapters/pgc"
	"gmart/internal/domain"
)

const sqlInsertIntoUsers = `
	INSERT INTO users (login, password_hash)
	VALUES ($1, $2)
	RETURNING id
`
const sqlSelectFromUsers = `
	SELECT id, password_hash
	FROM users
	WHERE login = $1
`
const sqlInsertIntoSessions = `
	INSERT INTO sessions (id, user_id, user_agent, ip_address, expires_at) VALUES ($1, $2, $3, $4, $5)
`
const sqlDeleteFromSessions = `
	DELETE FROM sessions
	WHERE id = $1
`
const sqlUpdateWithCheck = `
	UPDATE sessions
	SET expires_at = $2, ip_address = $3, user_agent = $4
	WHERE id = $1 AND expires_at > $5
	RETURNING user_id
`

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrSessionExpired    = errors.New("session expired")
	ErrSessionNotFound   = errors.New("session not found")
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=repository_mock.go  -package=user

type AuthMetrics interface {
	IncLoginError(reason string)
	ObserveBcrypt(d time.Duration)
}

type AuthRepo struct {
	pg          pgc.PgInstance
	metrics     AuthMetrics
	now         func() time.Time
	expDuration time.Duration
	avgBcrypt   atomic.Int64
}

func NewAuthRepo(pg pgc.PgInstance, expDuration time.Duration, m AuthMetrics) *AuthRepo {
	r := &AuthRepo{
		pg:          pg,
		metrics:     m,
		now:         time.Now,
		expDuration: expDuration,
	}
	r.avgBcrypt.Store((100 * time.Millisecond).Nanoseconds())
	return r
}

// ExpDuration время протухания сессионной куки
func (r *AuthRepo) ExpDuration() time.Duration {
	return r.expDuration
}

// SignUp регистрирует нового пользователя и сразу же создает сессию
func (r *AuthRepo) SignUp(ctx context.Context, login domain.Login, password domain.Password, userAgent domain.UserAgent, ip string) (domain.UserID, domain.SessionID, error) {
	start := r.now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if r.metrics != nil {
		r.metrics.ObserveBcrypt(r.now().Sub(start))
	}
	if err != nil {
		return 0, domain.SessionID(uuid.Nil), fmt.Errorf("hash password fail: %w", err)
	}

	// Создаем в БД пользователя и его сессионную куку
	var userID domain.UserID
	sessionID := domain.SessionID(uuid.New())
	err = r.pg.Tx(ctx, func(ctx context.Context, tx pgc.PgxTxIface) error {
		err := tx.QueryRow(ctx,
			sqlInsertIntoUsers,
			login,
			string(hashedPassword),
		).Scan(&userID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			sqlInsertIntoSessions,
			sessionID,
			userID,
			userAgent,
			ip,
			start.Add(r.expDuration),
		)
		return err
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// 23505 - код ошибки unique_violation в PostgreSQL
			if pgErr.Code == pgerrcode.UniqueViolation {
				return 0, domain.SessionID(uuid.Nil), ErrUserAlreadyExists
			}
		}
		return 0, domain.SessionID(uuid.Nil), fmt.Errorf("create user fail: %w", err)
	}

	return userID, sessionID, nil
}

// SignIn проверяет пароль и создает сессию (Refresh Token)
func (r *AuthRepo) SignIn(ctx context.Context, login domain.Login, password domain.Password, userAgent domain.UserAgent, ip string) (domain.UserID, domain.SessionID, error) {
	var userID domain.UserID
	var hash string

	// 1. Получаем хеш из БД
	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		return pool.QueryRow(ctx, sqlSelectFromUsers, login).Scan(&userID, &hash)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if r.metrics != nil {
				r.metrics.IncLoginError("err_user_not_found")
			}
			return 0, domain.SessionID(uuid.Nil), ErrUserNotFound
		}
		return 0, domain.SessionID(uuid.Nil), err
	}

	// 2. Сравниваем Bcrypt
	startBcrypt := r.now()
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if r.metrics != nil {
		r.metrics.ObserveBcrypt(r.now().Sub(startBcrypt))
	}
	if err != nil {
		if r.metrics != nil {
			r.metrics.IncLoginError("err_invalid_password")
		}
		return 0, domain.SessionID(uuid.Nil), ErrInvalidPassword
	}

	// 3. Создаем сессию
	sessionID := domain.SessionID(uuid.New())
	expiresAt := r.now().Add(r.expDuration)

	err = r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		_, err := pool.Exec(ctx, sqlInsertIntoSessions, sessionID, userID, userAgent, ip, expiresAt)
		return err
	})

	if err != nil {
		return 0, domain.SessionID(uuid.Nil), fmt.Errorf("create session fail: %w", err)
	}

	return userID, sessionID, nil
}

// CheckSession проверяет валидность Refresh Token
func (r *AuthRepo) RefreshSession(ctx context.Context, sessionID domain.SessionID, userAgent domain.UserAgent, ip string) (domain.UserID, error) {
	var userID domain.UserID
	currentTime := r.now()
	newExpiresAt := currentTime.Add(r.expDuration)

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		// UPDATE вернет id пользователя, только если сессия существует и НЕ просрочена
		return pool.QueryRow(ctx, sqlUpdateWithCheck, sessionID, newExpiresAt, ip, userAgent, currentTime).Scan(&userID)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Либо сессии нет, либо она уже протухла (не прошла условие expires_at > NOW())
			return 0, ErrSessionExpired
		}
		return 0, err
	}

	return userID, nil
}

// SignOut удаляет сессию
func (r *AuthRepo) SignOut(ctx context.Context, sessionID domain.SessionID) error {
	return r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		ct, err := pool.Exec(ctx, sqlDeleteFromSessions, sessionID)
		if ct.RowsAffected() == 0 {
			return ErrSessionNotFound
		}
		return err
	})
}

// SignInSleep — обертка для защиты от Timing Attack без лишней нагрузки на CPU
func (r *AuthRepo) SignInSleep(ctx context.Context, login domain.Login, password domain.Password, ua domain.UserAgent, ip string) (domain.UserID, domain.SessionID, error) {
	// Запрашиваем настоящую функцию
	start := r.now()
	uid, sid, err := r.SignIn(ctx, login, password, ua, ip)
	duration := r.now().Sub(start)
	avgBcrypt := r.avgBcrypt.Load()

	if err == nil || errors.Is(err, ErrInvalidPassword) {
		// Сработал bcrypt. Собираем статистику
		if avgBcrypt == 0 {
			r.avgBcrypt.Store(duration.Nanoseconds())
		} else {
			r.avgBcrypt.Store((avgBcrypt*7 + duration.Nanoseconds()*3) / 10)
		}
	} else if errors.Is(err, ErrUserNotFound) {
		// bcrypt не вызывался. Усложняем жизнь хакеру
		target := time.Duration(avgBcrypt)

		diff := target - duration
		if diff > 0 {
			jitter := diff / 7
			sleepTime := diff - jitter
			if jitter > 0 {
				n, _ := rand.Int(rand.Reader, big.NewInt(jitter.Nanoseconds()*2))
				sleepTime += time.Duration(n.Int64())
			}
			select {
			case <-time.After(sleepTime):
			case <-ctx.Done():
				return 0, domain.SessionID(uuid.Nil), ctx.Err()
			}
		}
	}

	return uid, sid, err
}
