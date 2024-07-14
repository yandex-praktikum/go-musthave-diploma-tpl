package user

import (
	"context"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	http2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http/middlware/private_router"
	user "github.com/GTech1256/go-musthave-diploma-tpl/internal/service/order"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/luhn"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"time"
)

type JWTClient interface {
	BuildJWTString(userId int) (string, error)
	GetTokenExp() time.Duration
	GetUserID(tokenString string) (int, error)
}

type UserExister interface {
	GetIsUserExistById(ctx context.Context, userId int) (bool, error)
}

type Service interface {
	Create(ctx context.Context, userId int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
}

type handler struct {
	logger      logging2.Logger
	service     Service
	jwtClient   JWTClient
	userExister UserExister
}

func NewHandler(logger logging2.Logger, updateService Service, jwtClient JWTClient, userExister UserExister) http2.Handler {
	return &handler{
		logger:      logger,
		service:     updateService,
		jwtClient:   jwtClient,
		userExister: userExister,
	}
}

func (h handler) Register(router *chi.Mux) {
	router.Post("/api/user/orders", privateRouter.WithPrivateRouter(http.HandlerFunc(h.orders), h.logger, h.jwtClient, h.userExister))
}

// orders /api/user/orders
// Возможные коды ответа:
// - `200` — номер заказа уже был загружен этим пользователем; +
// - `202` — новый номер заказа принят в обработку; +
// - `400` — неверный формат запроса; +
// - `401` — пользователь не аутентифицирован; (проверяется через мидлвару) +
// - `409` — номер заказа уже был загружен другим пользователем; +
// - `422` — неверный формат номера заказа; +
// - `500` — внутренняя ошибка сервера. +
func (h handler) orders(writer http.ResponseWriter, request *http.Request) {
	order, err := decodeOrder(request.Body)

	// `400` — неверный формат запроса;
	if err != nil {
		h.logger.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// `422` — неверный формат номера заказа;
	if !luhn.Valid(*order) {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userId := request.Context().Value("userId").(int)

	_, err = h.service.Create(request.Context(), userId, (*entity.OrderNumber)(order))
	if errors.Is(err, user.ErrOrderNumberAlreadyUploadByCurrentUser) {
		// `200` — номер заказа уже был загружен этим пользователем;
		writer.WriteHeader(http.StatusOK)
		return
	}
	if errors.Is(err, user.ErrOrderNumberAlreadyUploadByOtherUser) {
		// `409` — номер заказа уже был загружен другим пользователем;
		writer.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		// `500` — внутренняя ошибка сервера.
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// `202` — новый номер заказа принят в обработку;
	writer.WriteHeader(http.StatusAccepted)
}

func decodeOrder(body io.ReadCloser) (*int, error) {
	// Читаем тело запроса
	bodyByte, err := io.ReadAll(body)
	if err != nil {

		return nil, err
	}

	// Преобразовываем тело запроса в строку
	bodyString := string(bodyByte)

	// Пробуем преобразовать строку в число
	number, err := strconv.Atoi(bodyString)
	if err != nil {
		return nil, err
	}

	return &number, nil
}
