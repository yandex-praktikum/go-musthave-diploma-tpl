package composition

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	userHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user"
	userLoginHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/login"
	userRegisterHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/register"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	userRepository "github.com/GTech1256/go-musthave-diploma-tpl/internal/repository/user"
	userService "github.com/GTech1256/go-musthave-diploma-tpl/internal/service/user"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"time"
)

type JWTClient interface {
	BuildJWTString(userId int) (string, error)
	GetUserID(tokenString string) (int, error)
	GetTokenExp() time.Duration
}

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Service interface {
	Ping(ctx context.Context) error
	Register(ctx context.Context, userRegister *entity.UserRegisterJSON) (*entity.UserDB, error)
	Login(ctx context.Context, userRegister *entity.UserLoginJSON) (*entity.UserDB, error)
	GetIsUserExistById(ctx context.Context, userId int) (bool, error)
}

type UsersComposite struct {
	Service Service
	Handler http.Handler
}

func NewUserComposite(cfg *config.Config, logger logging.Logger, db DB, jwtClient JWTClient) (*UsersComposite, error) {
	storage := userRepository.NewStorage(db, logger)

	service := userService.NewUserService(logger, storage, cfg)

	handler := newUserHandler(logger, service, jwtClient)

	return &UsersComposite{
		Service: service,
		Handler: handler,
	}, nil
}

type UserHandler struct {
	logger    logging.Logger
	service   Service
	jwtClient JWTClient
}

func newUserHandler(logger logging.Logger, service Service, jwtClient JWTClient) http.Handler {
	return &UserHandler{
		logger:    logger,
		service:   service,
		jwtClient: jwtClient,
	}
}

func (h UserHandler) Register(router *chi.Mux) {
	handler1 := userHandler.NewHandler(h.logger, h.service)
	handler1.Register(router)

	handler2 := userRegisterHandler.NewHandler(h.logger, h.service, h.jwtClient)
	handler2.Register(router)

	handler3 := userLoginHandler.NewHandler(h.logger, h.service, h.jwtClient)
	handler3.Register(router)
}
