package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	http2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	privateRouter "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/middlware/private_router"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http/utils/auth"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/service/user"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/luhn"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"time"
)

type JWTClient interface {
	BuildJWTString(userID int) (string, error)
	GetTokenExp() time.Duration
	GetUserID(tokenString string) (int, error)
}

type UserExister interface {
	GetIsUserExistByIВ(ctx context.Context, userID int) (bool, error)
}

type Service interface {
	Login(ctx context.Context, userRegister *entity.UserLoginJSON) (*entity.UserDB, error)
	GetByID(ctx context.Context, userID int) (*entity.UserDB, error)
	Withdraw(ctx context.Context, withdrawalRawRecord entity.WithdrawalRawRecord) (*entity.UserDB, error)
	GetWithdrawals(ctx context.Context, userID int) ([]*entity.WithdrawalDB, error)
}

type handler struct {
	logger      logging2.Logger
	service     Service
	jwtClient   JWTClient
	userExister UserExister
}

type WithdrawRequestBody struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
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
	router.Get("/api/user/balance", privateRouter.WithPrivateRouter(http.HandlerFunc(h.userBalance), h.logger, h.jwtClient, h.userExister))
	router.Post("/api/user/balance/withdraw", privateRouter.WithPrivateRouter(http.HandlerFunc(h.userBalanceWithdraw), h.logger, h.jwtClient, h.userExister))
	router.Get("/api/user/withdrawals", privateRouter.WithPrivateRouter(http.HandlerFunc(h.userWithdrawals), h.logger, h.jwtClient, h.userExister))
}

// userBalance /api/user/balance
func (h handler) userBalance(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	userID := auth.GetUserIDFromContext(ctx)

	userDB, err := h.service.GetByID(ctx, *userID)

	if err != nil {
		h.logger.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	encodedUserBalance, err := encodeUserBalance(userDB)
	if err != nil {
		h.logger.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(encodedUserBalance)
}

func encodeUserBalance(userDB *entity.UserDB) ([]byte, error) {
	userBalance := &entity.UserBalanceJSON{
		Current:   userDB.Wallet,
		Withdrawn: userDB.Withdrawn,
	}

	return json.Marshal(userBalance)
}

// userBalance /api/user/balance
// Возможные коды ответа:
// - `200` — успешная обработка запроса;
// - `401` — пользователь не авторизован; - в мидлваре выше
// - `402` — на счету недостаточно средств;
// - `422` — неверный номер заказа;
// - `500` — внутренняя ошибка сервера.
func (h handler) userBalanceWithdraw(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := auth.GetUserIDFromContext(ctx)
	withdrawRequestBody, err := decodeWithdrawRequestBody(&request.Body)
	// `500` — внутренняя ошибка сервера.
	if err != nil {
		h.logger.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	orderNum, err := strconv.Atoi(withdrawRequestBody.Order)
	// `422` — неверный номер заказа;
	if err != nil || !luhn.Valid(orderNum) {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	withdrawalRawRecord := entity.WithdrawalRawRecord{
		Order:  withdrawRequestBody.Order,
		Sum:    withdrawRequestBody.Sum,
		UserID: *userID,
	}

	_, err = h.service.Withdraw(ctx, withdrawalRawRecord)
	// `402` — на счету недостаточно средств;
	if errors.Is(err, user.ErrWithdrawCountGreaterThanUserBalance) {
		h.logger.Info(err)
		writer.WriteHeader(http.StatusPaymentRequired)
		return
	}
	// `500` — внутренняя ошибка сервера.
	if err != nil {
		h.logger.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// `200` — успешная обработка запроса;
	writer.WriteHeader(http.StatusOK)
}

func decodeWithdrawRequestBody(body *io.ReadCloser) (*WithdrawRequestBody, error) {
	var withdrawRequestBody WithdrawRequestBody

	decoder := json.NewDecoder(*body)
	err := decoder.Decode(&withdrawRequestBody)
	if err != nil {
		return nil, err
	}

	return &withdrawRequestBody, nil
}

// userWithdrawals /api/user/balance/withdrawals
// Возможные коды ответа:
// - `200` — успешная обработка запроса.
// Формат ответа:
// `200 OK HTTP/1.1
//
//	Content-Type: application/json
//	[
//	    {
//	        "order": "2377225624",
//	        "sum": 500,
//	        "processed_at": "2020-12-09T16:09:57+03:00"
//	    }
//	]`
//
// - `204` — нет ни одного списания.
// - `401` — пользователь не авторизован.
// - `500` — внутренняя ошибка сервера.
func (h handler) userWithdrawals(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := auth.GetUserIDFromContext(ctx)

	withdrawals, err := h.service.GetWithdrawals(ctx, *userID)
	// `500` — внутренняя ошибка сервера.
	if err != nil {
		h.logger.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// `204` — нет ни одного списания.
	if len(withdrawals) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	withdrawalsJSONs := make([]entity.WithdrawalJSON, 0)

	for _, withdrawal := range withdrawals {
		withdrawalsJSON := entity.WithdrawalJSON{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt.Time.Format(time.RFC3339),
		}
		withdrawalsJSONs = append(withdrawalsJSONs, withdrawalsJSON)
	}

	withdrawalsByte, err := json.Marshal(withdrawalsJSONs)
	// `500` — внутренняя ошибка сервера.
	if err != nil {

		h.logger.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// `200` — успешная обработка запроса;
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(withdrawalsByte)
}
