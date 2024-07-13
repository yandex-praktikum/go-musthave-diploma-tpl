package composition

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	userHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user"
	userBalanceHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/balance"
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
	BuildJWTString(userID int) (string, error)
	GetUserID(tokenString string) (int, error)
	GetTokenExp() time.Duration
}

type UserExister interface {
	GetIsUserExistByIВ(ctx context.Context, userID int) (bool, error)
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
	GetIsUserExistByIВ(ctx context.Context, userID int) (bool, error)
	GetByID(ctx context.Context, userID int) (*entity.UserDB, error)
	Withdraw(ctx context.Context, withdrawalRawRecord entity.WithdrawalRawRecord) (*entity.UserDB, error)
	GetWithdrawals(ctx context.Context, userID int) ([]*entity.WithdrawalDB, error)
}

type Storage interface {
	IncrementBalance(ctx context.Context, userID int, incValue float64) (*entity.UserDB, error)
}

type UsersComposite struct {
	Service Service
	Storage Storage
	Handler http.Handler
}

func NewUserComposite(cfg *config.Config, logger logging.Logger, db DB, jwtClient JWTClient) (*UsersComposite, error) {
	storage := userRepository.NewStorage(db, logger)

	service := userService.NewUserService(logger, storage, cfg)

	handler := newUserHandler(logger, service, jwtClient, service)

	return &UsersComposite{
		Service: service,
		Storage: storage,
		Handler: handler,
	}, nil
}

type UserHandler struct {
	logger      logging.Logger
	service     Service
	jwtClient   JWTClient
	userExister UserExister
}

func newUserHandler(logger logging.Logger, service Service, jwtClient JWTClient, userExister UserExister) http.Handler {
	return &UserHandler{
		logger:      logger,
		service:     service,
		jwtClient:   jwtClient,
		userExister: userExister,
	}
}

func (h UserHandler) Register(router *chi.Mux) {
	userHandlerInstance := userHandler.NewHandler(h.logger, h.service)
	userHandlerInstance.Register(router)

	userRegisterHandlerInstance := userRegisterHandler.NewHandler(h.logger, h.service, h.jwtClient)
	userRegisterHandlerInstance.Register(router)

	userLoginHandlerInstance := userLoginHandler.NewHandler(h.logger, h.service, h.jwtClient)
	userLoginHandlerInstance.Register(router)

	userBalanceHandlerInstance := userBalanceHandler.NewHandler(h.logger, h.service, h.jwtClient, h.userExister)
	userBalanceHandlerInstance.Register(router)
}
