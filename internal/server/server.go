package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Raime-34/gophermart.git/internal/cfg"
	"github.com/Raime-34/gophermart.git/internal/cookie"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/gophermart"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Server struct {
	router        *chi.Mux
	gophermart    gophermartInterface
	cookieHandler cookieHandlerInterface
}

func newServer(dsn string) *Server {
	ctx := context.Background() // TODO

	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Fatal("Failed to create a config", zap.Error(err))
	}

	connPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		logger.Fatal("Error while creating connection to the database", zap.Error(err))
	}

	return &Server{
		gophermart:    gophermart.NewGophermart(ctx, connPool),
		cookieHandler: cookie.NewCookieHandler(),
	}
}

func (s *Server) mountHandlers() {
	s.router = chi.NewRouter()

	s.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.registerUser)
		r.Post("/login", s.loginUser)
	})
}

func StartServer() {
	config := cfg.GetConfig()

	migration(config.DbDSN)
	server := newServer(config.DbDSN)

	server.mountHandlers()

	http.ListenAndServe(config.Address, server.router)
}

func migration(dsn string) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("sql open", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("db ping", zap.Error(err))
	}

	if err := goose.Up(db, "/app//migrations"); err != nil {
		logger.Fatal("goose up", zap.Error(err))
	}
}

func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userCredential dto.UserCredential
	err := decoder.Decode(&userCredential)
	if err != nil {
		logger.Error("Failed to decode UserCredential: %v", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.gophermart.RegisterUser(r.Context(), userCredential)
	if err != nil {
		logger.Error("Failed to register user: %v", zap.Error(err))
		w.WriteHeader(http.StatusConflict)
		return
	}
}

func (s *Server) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userCredential dto.UserCredential
	err := decoder.Decode(&userCredential)
	if err != nil {
		logger.Error("Failed to decode UserCredential", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userData, err := s.gophermart.LoginUser(r.Context(), userCredential)
	if err != nil {
		if errors.Is(err, gophermart.ErrUserNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
		}
		if errors.Is(err, gophermart.ErrIncorrectPassword) {
			w.WriteHeader(http.StatusUnauthorized)
		}
		return
	}

	s.setCookie(w, userData)
}

func (s *Server) setCookie(w http.ResponseWriter, userData *dto.UserData) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	sid := hex.EncodeToString(b)

	s.cookieHandler.Set(sid, userData)
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: sid,
	})
}
