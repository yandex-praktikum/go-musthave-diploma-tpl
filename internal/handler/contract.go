package handler

import (
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type Autorisation interface {
	CreateUser(user models.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Orders interface {
	CreateOrder(num, user_id int, status string) (int, time.Time, error)
	GetOrders(user_id int) ([]models.Order, error)
}

type Storage struct {
	Autorisation
	Orders
}

func NewStorage(repos *repository.Repository) *Storage {
	return &Storage{
		Autorisation: NewAuthStorage(repos.Autorisation),
		Orders:       NewOrdersStorage(repos.Orders),
	}
}
