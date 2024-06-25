package application

import (
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/ShukinDmitriy/gophermart/internal/services"
	"gorm.io/gorm"
)

var App *Application

type Application struct {
	Conf                *config.Config
	DB                  *gorm.DB
	AccountRepository   *repositories.AccountRepository
	OperationRepository *repositories.OperationRepository
	OrderRepository     *repositories.OrderRepository
	UserRepository      *repositories.UserRepository
	AccrualService      *services.AccrualService
}

func NewApplication(
	conf *config.Config,
	DB *gorm.DB,
	accountRepository *repositories.AccountRepository,
	operationRepository *repositories.OperationRepository,
	orderRepository *repositories.OrderRepository,
	userRepository *repositories.UserRepository,
	accrualService *services.AccrualService,
) *Application {
	return &Application{
		Conf:                conf,
		DB:                  DB,
		AccountRepository:   accountRepository,
		OperationRepository: operationRepository,
		OrderRepository:     orderRepository,
		UserRepository:      userRepository,
		AccrualService:      accrualService,
	}
}

func AppFactory(DB *gorm.DB, conf *config.Config) {
	accountRepository := repositories.NewAccountRepository(DB)
	operationRepository := repositories.NewOperationRepository(DB)
	orderRepository := repositories.NewOrderRepository(DB)
	userRepository := repositories.NewUserRepository(DB)

	accrualService := services.NewAccrualService(
		conf.AccrualSystemAddress,
		accountRepository,
		operationRepository,
		orderRepository,
	)

	App = NewApplication(
		conf,
		DB,
		accountRepository,
		operationRepository,
		orderRepository,
		userRepository,
		accrualService,
	)
}
