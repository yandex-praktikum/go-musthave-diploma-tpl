package session

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

type Implementation struct {
	c *pgxpool.Pool
}

var sessionTable = `create table if not exists session (
    id    bigint GENERATED ALWAYS AS IDENTITY primary key,
    user_id       bigint NOT NULL,
    token 	      text NOT NULL,
    device 		  text NOT NULL,
    created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'))`

var tables = []string{
	sessionTable,
}

// NewImplementation ...
func NewImplementation(ctx context.Context, c *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := c.Exec(ctx, t)
		if err != nil {
			return nil, nil
		}
	}

	return &Implementation{
		c: c,
	}, nil
}

var jwtMethod = jwt.SigningMethodES256

func (repo *Implementation) SetSession(ctx context.Context, session domain.SessionInfo) (domain.SessionToken, error) {
	t := jwt.NewWithClaims(jwtMethod, jwt.MapClaims{
		"userID": session.UserID,
		"UUID":   uuid.New().String(),
	})

	token, err := t.SigningString()
	if err != nil {
		return domain.SessionToken{}, err
	}

	_, err = repo.c.Exec(ctx, `INSERT INTO session(token, user_id, device) values($1, $2, $3);`,
		token, session.UserID.ID, session.Device)
	if err != nil {
		return domain.SessionToken{}, err
	}

	return domain.SessionToken{
		Token: token,
	}, nil
}

type session struct {
	ID        sql.NullInt64  `db:"id"`
	UserID    sql.NullInt64  `db:"user_id"`
	Token     sql.NullString `db:"token"`
	Device    sql.NullString `db:"device"`
	CreatedAt sql.NullTime   `db:"created_at"`
}

func (repo *Implementation) GetSessions(ctx context.Context, token domain.SessionToken) (domain.SessionInfo, error) {
	s := session{}
	if err := repo.c.QueryRow(ctx, `Select id, user_id, created_at, device from session where token = $1`,
		token.Token).Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.Device); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SessionInfo{}, serviceerrors.NewNotFound().
				Wrap(domain.ErrNotFound, "session not found")
		}

		return domain.SessionInfo{}, err
	}

	return domain.SessionInfo{
		ID: domain.ID{
			ID: uint64(s.ID.Int64),
		},
		CreatedAt: s.CreatedAt.Time,
		UserID: domain.ID{
			ID: uint64(s.UserID.Int64),
		},
		Device: s.Device.String,
	}, nil
}
