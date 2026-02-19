package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/cfg"
	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/cookie"
	"github.com/Raime-34/gophermart.git/internal/gophermart"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "github.com/Raime-34/gophermart.git/docs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Server struct {
	router        *chi.Mux
	gophermart    gophermartInterface
	cookieHandler cookieHandlerInterface
}

func newServer(ctx context.Context, dsn string, wg *sync.WaitGroup) *Server {
	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Fatal("Failed to create a config", zap.Error(err))
	}

	connPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		logger.Fatal("Error while creating connection to the database", zap.Error(err))
	}

	return &Server{
		gophermart:    gophermart.NewGophermart(ctx, connPool, wg),
		cookieHandler: cookie.NewCookieHandler(),
	}
}

func (s *Server) mountHandlers() {
	s.router = chi.NewRouter()

	s.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.registerUser)
		r.Post("/login", s.loginUser)

		r.Group(func(r chi.Router) {
			r.Use(s.authMiddleware)

			r.Post("/orders", s.registerOrder)
			r.Get("/orders", s.getOrders)

			r.Post("/withdraw", s.registerWithdrawl)
			r.Get("/withdrawals", s.getWithdrawls)

			r.Get("/balance", s.getBalance)
		})
	})

	address := cfg.GetConfig().Address
	localhostPrefix := "localhost"
	if !strings.HasPrefix(address, localhostPrefix) {
		address = fmt.Sprintf("%v%v", localhostPrefix, address)
	}
	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%v/swagger/doc.json", address)),
	))
}

// Мидлвейр для проверки аутентификации
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil || c.Value == "" {
			logger.Error("Session id is not found in header")
			logger.Info(r.Header.Get("Cookie"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userData, ok := s.cookieHandler.Get(c.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), consts.UserIdKey, userData.Uuid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func StartServer(ctx context.Context, wg *sync.WaitGroup) {
	config := cfg.GetConfig()

	migration(config.DbDSN)
	server := newServer(ctx, config.DbDSN, wg)

	server.mountHandlers()

	go http.ListenAndServe(config.Address, server.router)
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
