package apiserver

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gorilla/sessions"
	"github.com/iRootPro/gophermart/internal/entity"
	"github.com/iRootPro/gophermart/internal/store"
	"github.com/iRootPro/gophermart/internal/store/sqlstore"
	"github.com/iRootPro/gophermart/internal/utils"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"time"
)

const sessionName = "gophermart_session"

var (
	errIncorrectLoginOrPassword = errors.New("incorrect login or password")
	errUserIsNotAuthorized      = errors.New("user is not authorized")
)

type APIServer struct {
	config       *Config
	logger       *logrus.Logger
	router       *echo.Echo
	store        *sqlstore.Store
	sessionStore sessions.Store
}

func NewAPIServer(config *Config, store *sqlstore.Store, sessionsStore sessions.Store) *APIServer {
	return &APIServer{
		config:       config,
		logger:       logrus.New(),
		router:       echo.New(),
		store:        store,
		sessionStore: sessionsStore,
	}
}

func (s *APIServer) Start() error {
	s.logger.Infof("starting api server on %s", s.config.RunAddress)

	s.configureRouter()

	if err := s.router.Start(s.config.RunAddress); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *APIServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutdown api server")
	err := s.router.Shutdown(ctx)
	if err != nil {
		s.logger.Fatal(err)
	}
	return s.router.Close()
}

func (s *APIServer) configureRouter() {
	userPublic := s.router.Group("/api/user/")
	userPublic.POST("register", s.handleUserCreate())
	userPublic.POST("login", s.handleUserLogin())

	//private
	user := s.router.Group("/api/user/", s.authUserMiddleware)
	user.POST("orders", s.authUserMiddleware(s.handleLoadOrders()))
	user.GET("orders", s.authUserMiddleware(s.handleListOrders()))
	user.GET("balance", s.authUserMiddleware(s.handleGetBalance()))
	user.POST("balance/withdraw", s.authUserMiddleware(s.handleWithdraw()))
	user.GET("withdrawals", s.authUserMiddleware(s.handleWithdrawals()))
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) handleUserCreate() echo.HandlerFunc {
	type request struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	return func(c echo.Context) error {
		req := &request{}

		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		u := &entity.User{
			Login:    req.Login,
			Password: req.Password,
		}

		if err := s.store.User().Create(u); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err := s.store.Balance().Create(u.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		session, err := s.sessionStore.Get(c.Request(), sessionName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		session.Values["user_id"] = u.ID
		if err = s.sessionStore.Save(c.Request(), c.Response(), session); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		u.Sanitize()
		return c.JSON(http.StatusOK, u)
	}
}

func (s *APIServer) handleUserLogin() echo.HandlerFunc {
	type request struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	return func(c echo.Context) error {
		req := &request{}

		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		u, err := s.store.User().FindByLogin(req.Login)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errIncorrectLoginOrPassword)
		}

		if !u.ComparePassword(req.Password) {
			return echo.NewHTTPError(http.StatusUnauthorized, errIncorrectLoginOrPassword)
		}

		session, err := s.sessionStore.Get(c.Request(), sessionName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		session.Values["user_id"] = u.ID
		if err = s.sessionStore.Save(c.Request(), c.Response(), session); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, nil)
	}
}

func (s *APIServer) handleLoadOrders() echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil || string(body) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		ok := utils.LuhnCheck(string(body))
		if !ok {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid card number")
		}

		userID := c.Get("user").(*entity.User).ID
		order := &entity.Order{
			UserID: userID,
			Number: string(body),
		}
		if err := s.store.Order().Create(order); err != nil {
			if err == store.ErrOrderNumberAlreadyExistInThisUser {
				return echo.NewHTTPError(http.StatusOK, err.Error())
			}

			if err == store.ErrOrderNumberAlreadyExistAnotherUser {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}

			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusAccepted, nil)

	}
}

func (s *APIServer) handleListOrders() echo.HandlerFunc {
	type responseItem struct {
		Number     string    `json:"number"`
		Status     string    `json:"status"`
		Accrual    float64   `json:"accrual"`
		UploadedAt time.Time `json:"uploaded_at"`
	}
	type response []responseItem

	return func(c echo.Context) error {
		userID := c.Get("user").(*entity.User).ID
		orders, err := s.store.Order().FindByUserID(userID)
		if err != nil {
			s.logger.Errorln("error while getting orders from db", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if len(orders) == 0 {
			return echo.NewHTTPError(http.StatusNoContent, "orders not found")
		}

		result := response{}
		for _, order := range orders {
			accrual, _ := order.Accrual.Float64()
			result = append(result, responseItem{
				Number:     order.Number,
				Status:     order.Status,
				Accrual:    accrual,
				UploadedAt: order.UploadedAt,
			})
		}

		return c.JSON(http.StatusOK, result)
	}
}

// Middlwewares
func (s *APIServer) authUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := s.sessionStore.Get(c.Request(), sessionName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		userID, ok := session.Values["user_id"]
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, errUserIsNotAuthorized)
		}

		u, err := s.store.User().FindByID(userID.(int))
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errUserIsNotAuthorized)

		}

		c.Set("user", u)
		return next(c)
	}
}

func (s *APIServer) handleGetBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user").(*entity.User).ID
		balance, err := s.store.Balance().Get(userID)
		type response struct {
			Current   float64 `json:"current"`
			Withdrawn float64 `json:"withdrawn"`
		}
		if err != nil {
			if err == sql.ErrNoRows {
				emptyBalance := &response{
					Current:   0,
					Withdrawn: 0,
				}
				return c.JSON(http.StatusOK, emptyBalance)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		current, _ := balance.Current.Float64()
		withdrawn, _ := balance.Withdrawn.Float64()

		responseBalance := response{
			Current:   current,
			Withdrawn: withdrawn,
		}

		return c.JSON(http.StatusOK, responseBalance)
	}
}

func (s *APIServer) handleWithdraw() echo.HandlerFunc {
	return func(c echo.Context) error {
		type response struct {
			Order string  `json:"order"`
			Sum   float64 `json:"sum"`
		}
		res := &response{}
		err := c.Bind(&res)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		userID := c.Get("user").(*entity.User).ID
		balance, err := s.store.Balance().Get(userID)
		if err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusUnprocessableEntity, "balance is empty")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		current, _ := balance.Current.Float64()
		if current < res.Sum {
			return echo.NewHTTPError(http.StatusPaymentRequired, "insufficient funds")
		}

		withdrawn, _ := balance.Withdrawn.Float64()
		newCurrent := current - res.Sum
		newWithdrawn := withdrawn + res.Sum

		if err := s.store.Balance().UpdateWithdrawnAndCurrentByUserID(userID, newCurrent, newWithdrawn); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err := s.store.Withdrawal().Create(userID, res.Order, res.Sum); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, nil)
	}
}

func (s *APIServer) handleWithdrawals() echo.HandlerFunc {
	type responseItem struct {
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}
	type response []responseItem

	return func(c echo.Context) error {
		userID := c.Get("user").(*entity.User).ID
		withdrawals, err := s.store.Withdrawal().GetByUserID(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if len(withdrawals) == 0 {
			return echo.NewHTTPError(http.StatusNoContent, "withdrawals not found")
		}

		result := response{}
		for _, withdrawal := range withdrawals {
			result = append(result, responseItem{
				Order:       withdrawal.OrderID,
				Sum:         withdrawal.Sum,
				ProcessedAt: withdrawal.ProcessedAt,
			})
		}

		return c.JSON(http.StatusOK, result)
	}
}
