package services

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/storage"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

// OrderService предоставляет методы для работы с заказами
type OrderService struct {
	orderRepo   repository.OrderRepository
	balanceRepo repository.BalanceRepository
}

// NewOrderService создаёт новый сервис заказов
func NewOrderService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
	}
}

// UploadOrder загружает номер заказа для пользователя
// Возвращает: exists (уже загружен этим пользователем), conflict (загружен другим), error
func (s *OrderService) UploadOrder(ctx context.Context, userID int64, number string) (bool, bool, error) {
	// Валидируем номер заказа
	if !ValidateOrderNumber(number) {
		return false, false, ErrInvalidOrderNumber
	}

	// Проверяем, загружал ли этот пользователь данный номер
	_, err := s.orderRepo.GetByUserAndNumber(ctx, userID, number)
	if err == nil {
		// Заказ уже загружен этим пользователем
		return true, false, nil
	}
	if !errors.Is(err, repository.ErrOrderNotFound) {
		return false, false, err
	}

	// Проверяем, загружал ли другой пользователь данный номер
	existingOrderByNumber, err := s.orderRepo.GetByNumber(ctx, number)
	if err == nil && existingOrderByNumber.UserID != userID {
		// Заказ загружен другим пользователем
		return false, true, nil
	}

	// Создаём новый заказ
	_, err = s.orderRepo.Create(ctx, userID, number)
	if err != nil {
		if errors.Is(err, repository.ErrOrderAlreadyExists) {
			// Гонка: заказ уже создан другим процессом
			return false, true, nil
		}
		return false, false, err
	}

	return false, false, nil
}

// GetOrdersByUser получает все заказы пользователя
func (s *OrderService) GetOrdersByUser(ctx context.Context, userID int64) ([]storage.Order, error) {
	return s.orderRepo.GetByUser(ctx, userID)
}

// ValidateOrderNumber проверяет номер заказа по алгоритму Луна
func ValidateOrderNumber(number string) bool {
	if number == "" {
		return false
	}

	// Проверяем, что номер состоит только из цифр
	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}

	// Алгоритм Луна (опционально, но рекомендуется)
	return luhnCheck(number)
}

// luhnCheck реализует алгоритм Луна для проверки номера
func luhnCheck(number string) bool {
	digits := make([]int, len(number))
	for i, r := range number {
		digits[i] = int(r - '0')
	}

	sum := 0
	double := false

	// Проходим с конца
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}

	return sum%10 == 0
}

// ParseOrderNumber парсит номер заказа из строки
func ParseOrderNumber(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ErrInvalidOrderNumber
	}
	return s, nil
}

// GetBalance получает баланс пользователя
func (s *OrderService) GetBalance(ctx context.Context, userID int64) (*storage.UserBalance, error) {
	return s.balanceRepo.GetBalance(ctx, userID)
}

// WithdrawRequest запрос на списание средств
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Withdraw списывает средства со счёта пользователя
func (s *OrderService) Withdraw(ctx context.Context, userID int64, req WithdrawRequest) error {
	// Валидируем номер заказа
	if !ValidateOrderNumber(req.Order) {
		return ErrInvalidOrderNumber
	}

	// Атомарно обрабатываем списание (транзакция включает проверку баланса,
	// добавление записи в transactions и создание записи в withdrawals)
	if err := s.balanceRepo.ProcessWithdraw(ctx, userID, req.Order, req.Sum); err != nil {
		return err
	}

	return nil
}

// GetWithdrawalsByUser получает все списания пользователя
func (s *OrderService) GetWithdrawalsByUser(ctx context.Context, userID int64) ([]storage.Withdrawal, error) {
	return s.balanceRepo.GetWithdrawalsByUser(ctx, userID)
}

// ProcessOrderAccrual обрабатывает начисление по заказу
func (s *OrderService) ProcessOrderAccrual(ctx context.Context, orderID int64, userID int64, status string, accrual *float64, orderNumber string) error {
	return s.balanceRepo.ProcessOrderAccrual(ctx, orderID, userID, status, accrual, orderNumber)
}

// GetOrdersByStatus получает заказы по статусу
func (s *OrderService) GetOrdersByStatus(ctx context.Context, status string, limit int) ([]storage.Order, error) {
	return s.orderRepo.GetOrdersByStatus(ctx, status, limit)
}

// UpdateOrderStatus обновляет статус заказа
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status string, accrual *float64) error {
	return s.orderRepo.UpdateStatus(ctx, orderID, status, accrual)
}

// IntToPointerInt64 конвертирует int в *int64
func IntToPointerInt64(i int) *int64 {
	v := int64(i)
	return &v
}

// ParseAccrual парсит accrual из интерфейса в int64
func ParseAccrual(accrual any) (*int64, error) {
	if accrual == nil {
		return nil, nil
	}

	switch v := accrual.(type) {
	case int:
		return IntToPointerInt64(v), nil
	case int64:
		return &v, nil
	case float64:
		iv := int64(v)
		return &iv, nil
	case string:
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		return &iv, nil
	default:
		return nil, errors.New("invalid accrual type")
	}
}
