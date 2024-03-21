package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SmoothWay/gophermart/internal/model"
	"github.com/google/uuid"
)

const (
	accrualAPIURL = "%s/api/orders/%s"
)

var (
	ErrNoContent                    = errors.New("no content")
	ErrOrderAlreadyExistThisUser    = errors.New("order number already exist for this user")
	ErrOrderAlreadyExistAnotherUser = errors.New("order number already exist for another user")
)

type Storage interface {
	AddUser(login, password string) error
	GetUser(login, password string) (*model.User, error)
	AddOrder(userID uuid.UUID, order model.Order) error
	GetOrder(userID uuid.UUID, orderNumber string) (*model.Order, error)
	GetOrders(userID uuid.UUID) ([]model.Order, error)
	WithdrawalRequest(userID uuid.UUID, orderNumber string, sum float64) error
	GetBalance(uuid.UUID) (float64, float64, error)
	GetWithdrawals(userID uuid.UUID) ([]model.Withdrawal, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	storage    Storage
	client     HTTPClient
	secret     []byte
	accrualURL string
}

func New(storage Storage, client HTTPClient, secret []byte, url string) *Service {
	return &Service{
		storage:    storage,
		secret:     secret,
		accrualURL: url,
		client:     client,
	}
}

func (s *Service) GetWithdrawals(userID uuid.UUID) ([]model.Withdrawal, error) {
	return s.storage.GetWithdrawals(userID)
}

func (s *Service) GetBalance(userID uuid.UUID) (float64, float64, error) {
	return s.storage.GetBalance(userID)
}

func (s *Service) GetOrders(userID uuid.UUID) ([]model.Order, error) {
	return s.storage.GetOrders(userID)
}

func (s *Service) WithdrawalRequest(userID uuid.UUID, orderNumber string, sum float64) error {
	return s.storage.WithdrawalRequest(userID, orderNumber, sum)
}

func (s *Service) UploadOrder(userID uuid.UUID, orderNumber string) error {
	_, err := s.storage.GetOrder(userID, orderNumber)
	if err == nil {
		return ErrOrderAlreadyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var order model.Order
	o, err := s.fetchOrder(orderNumber)
	if err != nil {
		if errors.Is(err, ErrNoContent) {
			order = model.Order{
				Number: orderNumber,
				Status: "NEW",
			}

			return s.storage.AddOrder(userID, order)
		}

		return err
	}

	return s.storage.AddOrder(userID, *o)
}

func (s *Service) RegisterUser(login, password string) error {
	return s.storage.AddUser(login, password)
}

func (s *Service) Authenticate(login, password string) (string, error) {
	user, err := s.storage.GetUser(login, password)
	if err != nil {
		return "", err
	}

	return s.generateAccessToken(user.ID)
}

func (s *Service) generateAccessToken(id uuid.UUID) (string, error) {
	token := jwt.New()
	now := time.Now()
	token.Set(jwt.SubjectKey, id.String())
	token.Set(jwt.IssuedAtKey, now.Unix())
	token.Set(jwt.ExpirationKey, now.Add(10*time.Minute))
	signedToken, err := jwt.Sign(token, jwa.HS256, s.secret)
	if err != nil {
		return "", err
	}

	return string(signedToken), nil
}

func (s *Service) fetchOrder(orderNumber string) (*model.Order, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(accrualAPIURL, s.accrualURL, orderNumber), nil)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNoContent {
			return nil, ErrNoContent
		}
		return nil, fmt.Errorf("failed to fetch order info, status code %d", res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var o model.Order
	if err := json.Unmarshal(b, &o); err != nil {
		return nil, err
	}

	return &o, nil
}
