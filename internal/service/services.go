package service

import "github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"

// Services — контейнер всех сервисов приложения.
type Services struct {
	User    *UserService
	Order   *OrderService
	Balance *BalanceService
}

// ServicesOption — функция для настройки Services.
type ServicesOption func(*Services)

// WithUserService переопределяет сервис пользователей (полезно для тестов).
func WithUserService(svc *UserService) ServicesOption {
	return func(s *Services) {
		s.User = svc
	}
}

// WithOrderService переопределяет сервис заказов (полезно для тестов).
func WithOrderService(svc *OrderService) ServicesOption {
	return func(s *Services) {
		s.Order = svc
	}
}

// WithBalanceService переопределяет сервис баланса (полезно для тестов).
func WithBalanceService(svc *BalanceService) ServicesOption {
	return func(s *Services) {
		s.Balance = svc
	}
}

// NewServices создаёт все сервисы из контейнера репозиториев.
// Сервисы зависят от интерфейса RepositoryContainer, а не от конкретной реализации Storage.
func NewServices(container repository.RepositoryContainer, opts ...ServicesOption) *Services {
	s := &Services{
		User:    NewUserService(container.UserRepository()),
		Order:   NewOrderService(container.OrderRepository()),
		Balance: NewBalanceService(container.OrderRepository(), container.WithdrawalRepository()),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
