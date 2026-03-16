package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/NailUsmanov/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type DataBaseStorage struct {
	db *sql.DB
}

type Order struct {
	Number     string    `json:"number"`
	Status     *string   `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func HashPassword(password string) string {

	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := sha256.New()
	// передаём байты для хеширования
	h.Write([]byte(password))
	// вычисляем хеш
	return hex.EncodeToString(h.Sum(nil))

}

func NewDataBaseStorage(dsn string) (*DataBaseStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	//Настройка миграций
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver: %w", err)
	}

	// Инициализация мигратора
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise migrate driver: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	return &DataBaseStorage{db: db}, nil
}

func (d *DataBaseStorage) Registration(ctx context.Context, login, password string) error {
	// Проверим, нет ли пользователя уже в базе
	_, err := d.db.ExecContext(ctx, RegistrationPostgres, login, password)
	if err != nil {
		// Проверяем, не ошибка ли это из-за нарушения ограничения UNIQUE
		if strings.Contains(err.Error(), "duplicate key") {
			return ErrOrderAlreadyUsed
		}
		return fmt.Errorf("failed to save new user: %v", err)
	}
	return nil
}

func (d *DataBaseStorage) GetUserByLogin(ctx context.Context, login string) (string, error) {
	var hashedPassword string
	err := d.db.QueryRowContext(ctx, CheckLoginPostgres, login).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user: %v", err)
	}
	return hashedPassword, nil
}

func (d *DataBaseStorage) GetUserIDByLogin(ctx context.Context, login string) (int, error) {
	var userID int
	err := d.db.QueryRowContext(ctx, LoginIDPostgres, login).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}
	return userID, nil
}

func (d *DataBaseStorage) CheckHashMatch(ctx context.Context, login, password string) error {
	var usersHash string
	err := d.db.QueryRowContext(ctx, CheckHashPasswordPostgres, login).Scan(&usersHash)
	if err == sql.ErrNoRows {
		return fmt.Errorf("don't have userHash: %v", err)
	}
	if err != nil {
		return err
	}
	if usersHash != HashPassword(password) {
		return fmt.Errorf("invalid password hash")
	}
	return nil
}

func (d *DataBaseStorage) CreateNewOrder(ctx context.Context, userNumber int, numberOrder string, sugar *zap.SugaredLogger) error {
	sugar.Infof(">>> Creating order: order=%s, userID=%d", numberOrder, userNumber)
	fmt.Printf(">>> DEBUG: orderNum=%s, userID=%d\n", numberOrder, userNumber)
	_, err := d.db.ExecContext(ctx, CreateNewOrderPostgres, numberOrder, userNumber)
	if err != nil {
		fmt.Printf(">>> DEBUG CreateNewOrder error: %v\n", err)
		if strings.Contains(err.Error(), "duplicate key") {
			fmt.Println(">>> DEBUG Detected duplicate key violation")
			return ErrOrderAlreadyUploaded
		}
		fmt.Printf(">>> DEBUG CreateNewOrder error: %v\n", err)
		return fmt.Errorf("failed to insert new order: %w", err)
	}
	fmt.Println(">>> DEBUG CreateNewOrder completed successfully")
	return nil
}

func (d *DataBaseStorage) CheckExistOrder(ctx context.Context, numberOrder string) (bool, int, error) {
	var existingUserID int
	err := d.db.QueryRowContext(ctx, CheckUserOrderPostgres, numberOrder).Scan(&existingUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, -1, nil
		}
		return false, -1, fmt.Errorf("failed to check order existence: %w", err)
	}

	return true, existingUserID, nil
}

func (d *DataBaseStorage) GetOrdersByUserID(ctx context.Context, userID int) ([]Order, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// Создаем массив структур Order
	orders := make([]Order, 0)
	// Выполняем запрос в базу данных orders
	rows, err := d.db.QueryContext(ctx, GetUserOrdersQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("db query: %v", err)
	}
	defer rows.Close()
	// Сканируем полученные значения в массив структур Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan row: %v", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return orders, nil
}

func (d *DataBaseStorage) GetOrdersForAccrualUpdate(ctx context.Context) ([]Order, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// Создаем массив структур Order
	orders := make([]Order, 0)
	// Выполняем запрос в БД orders
	rows, err := d.db.QueryContext(ctx, GetOrdersForAccrual)
	if err != nil {
		return nil, fmt.Errorf("db query: %v", err)
	}
	defer rows.Close()
	// Сканируем полученные значения в массив структур orders
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.Number, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan row: %v", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration: %v", err)
	}
	return orders, nil
}

func (d *DataBaseStorage) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Выполняем обновление БД orders
	_, err := d.db.ExecContext(ctx, UpdateOrderStatusPostgres, status, accrual, number)
	if err != nil {
		return fmt.Errorf("exec row: %v", err)
	}
	return nil
}

func (d *DataBaseStorage) GetUserBalance(ctx context.Context, userID int) (current, withdrawn float64, err error) {
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	default:
	}
	var result sql.NullFloat64
	err = d.db.QueryRowContext(ctx, GetBalanceIncome, userID).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("failed scan query row: %v", err)
	}
	withdrawns, err := d.GetUserWithDrawns(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	total := result.Float64 - withdrawns
	return total, withdrawn, nil
}

func (d *DataBaseStorage) GetUserWithDrawns(ctx context.Context, userID int) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	var result sql.NullFloat64
	err := d.db.QueryRowContext(ctx, GetBalanceWithDrawn, userID).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed scan query row: %v", err)
	}
	return result.Float64, nil
}

func (d *DataBaseStorage) AddWithdrawOrder(ctx context.Context, userID int, orderNumber string, sum float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ok, _, err := d.CheckExistOrder(ctx, orderNumber)
	if err != nil {
		return fmt.Errorf("method CheckExistOrder failed: %v", err)
	}
	if ok {
		return ErrOrderAlreadyUsed
	}
	currentBalance, _, err := d.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("method GetUserBalance failed: %v", err)
	}
	if sum > currentBalance {
		return ErrNotEnoughFunds
	}
	_, err = d.db.ExecContext(ctx, AddWithdrawOrderPostgres, userID, orderNumber, sum)
	if err != nil {
		return fmt.Errorf("failed to update table orders: %v", err)
	}
	return nil
}

func (d *DataBaseStorage) GetAllUserWithdrawals(ctx context.Context, userID int) ([]models.UserWithDraw, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// Создаем массив структур для ответа
	allWithDrawls := make([]models.UserWithDraw, 0)
	// Выполняем запрос в БД orders, чтобы достать все позиции по конкретному Юзеру
	rows, err := d.db.QueryContext(ctx, GetAllWithDrawals, userID)
	if err != nil {
		return nil, fmt.Errorf("db query: %v", err)
	}
	defer rows.Close()
	// Сканируем полученные значения в массив структур
	for rows.Next() {
		var order models.UserWithDraw
		if err := rows.Scan(&order.NumberOrder, &order.Sum, &order.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan row: %v", err)
		}
		allWithDrawls = append(allWithDrawls, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return allWithDrawls, nil
}
