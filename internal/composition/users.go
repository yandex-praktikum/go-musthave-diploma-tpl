package composition

import (
	"context"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	sql2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/db/sql"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	userHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user"
	userLoginHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/login"
	userRegisterHandler "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/register"
	userRepository "github.com/GTech1256/go-musthave-diploma-tpl/internal/repository/user"
	userService "github.com/GTech1256/go-musthave-diploma-tpl/internal/service/user"
	jwt2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/jwt"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"time"
)

type Storage interface {
	Ping(ctx context.Context) error
}

type JWTClient interface {
	BuildJWTString(userId int) (string, error)
	GetUserID(tokenString string) (int, error)
	GetTokenExp() time.Duration
}

type Service interface {
	Ping(ctx context.Context) error
	Register(ctx context.Context, userRegister *entity.UserRegisterJSON) (*entity.UserDB, error)
	Login(ctx context.Context, userRegister *entity.UserLoginJSON) (*entity.UserDB, error)
}

type UsersComposite struct {
	Storage Storage
	Service Service
	Handler http.Handler
}

var (
	ErrNoSQLConnection = errors.New("нет подключения к БД")
)

func NewUserComposite(cfg *config.Config, logger logging.Logger) (*UsersComposite, error) {
	sql, err := sql2.NewSQL(*cfg.DatabaseURI)
	//defer sql.DB.Close()
	if err != nil {
		logger.Error(err)
		return nil, ErrNoSQLConnection
	}

	jwtClient := jwt2.NewJwt(*cfg.JWTTokenExp, *cfg.JWTSecretKey)
	storage := userRepository.NewStorage(sql.DB, logger)

	service := userService.NewUserService(logger, storage, cfg)

	handler := newHandler(logger, service, jwtClient)

	return &UsersComposite{
		Storage: storage,
		Service: service,
		Handler: handler,
	}, nil
}

type Handler struct {
	logger    logging.Logger
	service   Service
	jwtClient JWTClient
}

func newHandler(logger logging.Logger, service Service, jwtClient JWTClient) http.Handler {
	return &Handler{
		logger:    logger,
		service:   service,
		jwtClient: jwtClient,
	}
}

func (h Handler) Register(router *chi.Mux) {
	handler1 := userHandler.NewHandler(h.logger, h.service)
	handler1.Register(router)

	handler2 := userRegisterHandler.NewHandler(h.logger, h.service, h.jwtClient)
	handler2.Register(router)

	handler3 := userLoginHandler.NewHandler(h.logger, h.service, h.jwtClient)
	handler3.Register(router)
}
