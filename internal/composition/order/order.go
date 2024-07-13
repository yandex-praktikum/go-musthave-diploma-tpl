package composition

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual"
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
	Create(ctx context.Context, userID int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
	StartProcessingOrders()
	GetOrdersStatusJSONs(ctx context.Context, userID int) ([]*entity.OrderStatusJSON, error)
}

type UserService interface {
	GetIsUserExistByIВ(ctx context.Context, userID int) (bool, error)
}

type UserStorage interface {
	IncrementBalance(ctx context.Context, userID int, incValue float64) (*entity.UserDB, error)
}

type UsersComposite struct {
	Handler http.Handler
	Service Service
}

func NewOrderComposite(cfg *config.Config, logger logging.Logger, db DB, jwtClient JWTClient, userService UserService, userStorage UserStorage) (*UsersComposite, error) {
	storage := orderRepository.NewStorage(db, logger)
	accrualClient := accrual.New(*cfg.AccrualSystemAddress, logger)
	service := order.NewOrderService(accrualClient, logger, storage, cfg, userStorage)

	handler := newHandler(logger, service, jwtClient, userService)

	return &UsersComposite{
		Handler: handler,
		Service: service,
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
	userOrdersHandlerInstance := userOrdersHandler.NewHandler(h.logger, h.service, h.jwtClient, h.userExister)
	userOrdersHandlerInstance.Register(router)
}
