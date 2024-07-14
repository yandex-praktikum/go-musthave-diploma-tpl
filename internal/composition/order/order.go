package composition

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	userOrdersHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/orders"
	orderRepository "github.com/GTech1256/go-musthave-diploma-tpl/internal/repository/order"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/service/order"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type JWTClient interface {
	BuildJWTString(userId int) (string, error)
	GetUserID(tokenString string) (int, error)
	GetTokenExp() time.Duration
}

type UserExister interface {
	GetIsUserExistById(ctx context.Context, userId int) (bool, error)
}

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Service interface {
	Create(ctx context.Context, userId int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
}

type UserService interface {
	GetIsUserExistById(ctx context.Context, userId int) (bool, error)
}

type UsersComposite struct {
	Handler http.Handler
}

func NewOrderComposite(cfg *config.Config, logger logging.Logger, db DB, jwtClient JWTClient, userService UserService) (*UsersComposite, error) {
	storage := orderRepository.NewStorage(db, logger)

	service := user.NewOrderService(logger, storage, cfg)

	handler := newHandler(logger, service, jwtClient, userService)

	return &UsersComposite{
		Handler: handler,
	}, nil
}

type Handler struct {
	logger      logging.Logger
	service     Service
	jwtClient   JWTClient
	userExister UserExister
}

func newHandler(logger logging.Logger, service Service, jwtClient JWTClient, userExister UserExister) http.Handler {
	return &Handler{
		logger:      logger,
		service:     service,
		jwtClient:   jwtClient,
		userExister: userExister,
	}
}

func (h Handler) Register(router *chi.Mux) {
	handler1 := userOrdersHandler.NewHandler(h.logger, h.service, h.jwtClient, h.userExister)
	handler1.Register(router)
}
