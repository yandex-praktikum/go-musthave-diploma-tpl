package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"

	"github.com/kdv2001/loyalty/internal/domain"
)

type Implementation struct {
	c *pgxpool.Pool
}

var usersTable = `create table if not exists users (	
    	id    bigint GENERATED ALWAYS AS IDENTITY primary key,
     	state varchar NOT NULL                          
    	)`

var authTable = `create table if not exists auth (
    id            bigint GENERATED ALWAYS AS IDENTITY primary key,
    user_id       bigint NOT NULL, FOREIGN KEY (user_id)  REFERENCES users (id),
    login         varchar NOT NULL UNIQUE,
    password_hash varchar NOT NULL,
    salt          varchar NOT NULL,
    state varchar NOT NULL,
    created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'))`

var tables = []string{
	usersTable,
	authTable,
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

type auth struct {
	Id           sql.NullInt64  `db:"id"`
	UserID       sql.NullInt64  `db:"user_id"`
	Login        sql.NullString `db:"login"`
	PasswordHash sql.NullString `db:"password_hash"`
	Salt         sql.NullString `db:"salt"`
	State        sql.NullString `db:"state"`
	CreatedAt    sql.NullTime   `db:"created_at"`
}

// Argon2Config holds the parameters for Argon2 hashing.
type Argon2Config struct {
	TimeCost    uint32
	MemoryCost  uint32
	Parallelism uint8
	KeyLength   uint32
	SaltLength  uint32
}

var config = &Argon2Config{
	TimeCost:    3,         // Iterations
	MemoryCost:  64 * 1024, // 64MB
	Parallelism: 4,         // Threads
	KeyLength:   32,        // Length of the derived key
	SaltLength:  16,        // Length of the salt
}

const saltLength = 32

func generateSalt(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (repo *Implementation) Register(ctx context.Context, a domain.Login) (domain.ID, error) {
	salt, err := generateSalt(saltLength)
	if err != nil {
		return domain.ID{}, err
	}

	hashPassword := argon2.IDKey([]byte(a.Password), salt,
		config.TimeCost, config.MemoryCost, config.Parallelism, config.KeyLength)

	tx, err := repo.c.Begin(ctx)
	if err != nil {
		return domain.ID{}, err
	}

	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedHashPassword := base64.StdEncoding.EncodeToString(hashPassword)

	defer func() {
		if err != nil {
			// TODO  обработать ошибку
			_ = tx.Rollback(ctx)
		}
	}()

	uID := &sql.NullInt64{}
	err = tx.QueryRow(ctx, `INSERT INTO users(state) values($1) RETURNING id;`, "unknown").Scan(uID)
	if err != nil {
		return domain.ID{}, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO auth(login, user_id, password_hash, salt,  state) 
      values($1, $2, $3, $4, $5);`, a.Login, uID.Int64, encodedHashPassword, encodedSalt, "verified")
	if err != nil {
		return domain.ID{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.ID{}, err
	}

	return domain.ID{
		ID: uint64(uID.Int64),
	}, nil
}

func (repo *Implementation) Login(ctx context.Context, authReq domain.Login) (domain.Auth, error) {
	a := auth{}
	err := repo.c.QueryRow(ctx, `Select id, user_id, login, password_hash, salt from auth where login = $1`, authReq.Login).
		Scan(&a.Id, &a.UserID, &a.Login, &a.PasswordHash, &a.Salt)
	if err != nil {
		return domain.Auth{}, err
	}

	decodedSalt, err := base64.StdEncoding.DecodeString(a.Salt.String)
	if err != nil {
		return domain.Auth{}, err
	}

	hashPassword := argon2.IDKey([]byte(authReq.Password), decodedSalt,
		config.TimeCost, config.MemoryCost, config.Parallelism, config.KeyLength)

	encodeHashPassword := base64.StdEncoding.EncodeToString(hashPassword)

	if encodeHashPassword != a.PasswordHash.String {
		return domain.Auth{}, errors.New("bad password")
	}

	return domain.Auth{
		ID: domain.ID{},
		UserID: domain.ID{
			ID: uint64(a.UserID.Int64),
		},
		Login: authReq,
	}, nil
}
