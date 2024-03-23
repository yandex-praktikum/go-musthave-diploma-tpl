package api

import (
	"compress/flate"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/api/middleware"
	"github.com/A-Kuklin/gophermart/internal/config"
	"github.com/A-Kuklin/gophermart/internal/usecases"
)

type API struct {
	Usecases *usecases.UseCases
	logger   logrus.FieldLogger
}

func NewAPI(logger logrus.FieldLogger, uc *usecases.UseCases) *API {
	return &API{
		Usecases: uc,
		logger:   logger,
	}
}

func (a *API) SetUpRoutes(cfg *config.Config) *gin.Engine {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		a.logger.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	r := gin.New()
	r.Use(gzip.Gzip(flate.BestCompression))
	r.Use(middleware.Logger(a.logger))

	authMW := middleware.NewAuth(cfg, a.logger)

	apiGroup := r.Group("/api")

	apiUsers := apiGroup.Group("/user")

	apiUsers.POST("/register", a.CreateUser)
	apiUsers.POST("/login", a.Login)

	apiUsers.POST("/orders", authMW.TokenAccess, a.CreateOrder)
	apiUsers.GET("/orders", authMW.TokenAccess, a.GetOrders)
	apiUsers.GET("/balance", authMW.TokenAccess, a.GetBalance)
	apiUsers.POST("/balance/withdraw", authMW.TokenAccess, a.Withdraw)
	apiUsers.GET("/withdrawals", authMW.TokenAccess, a.GetWithdrawals)

	return r
}
