package storage

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDatabaseURI = "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"

var dbAvailable bool

func TestMain(m *testing.M) {
	storage, err := NewDatabaseStorage(context.Background(), testDatabaseURI)
	if err != nil {
		os.Exit(0) // База недоступна — пропускаем все тесты
	}
	defer storage.Close()
	dbAvailable = true
	os.Exit(m.Run())
}

// TestDatabaseStorage_Integration тестирует работу с реальной базой данных
func TestDatabaseStorage_Integration(t *testing.T) {
	if !dbAvailable {
		t.Skip("Database not available, skipping test")
	}
	storage, err := NewDatabaseStorage(context.Background(), testDatabaseURI)
	if err != nil {
		t.Skipf("Skipping database tests: failed to connect to database: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Проверяем соединение
	err = storage.Ping(ctx)
	require.NoError(t, err)

	// Очищаем базу данных перед тестами
	cleanupDatabase(t, storage)

	t.Run("CreateUser", func(t *testing.T) {
		// Создаем пользователя
		user, err := storage.CreateUser(ctx, "testuser", "hashedpassword")
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "hashedpassword", user.Password)
	})

	t.Run("GetUserByLogin", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Login)

		// Проверяем несуществующего пользователя
		user, err = storage.GetUserByLogin(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("CreateOrder", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Создаем заказ
		order, err := storage.CreateOrder(ctx, user.ID, "12345678903")
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, user.ID, order.UserID)
		assert.Equal(t, "12345678903", order.Number)
		assert.Equal(t, "NEW", order.Status)
	})

	t.Run("GetOrderByNumber", func(t *testing.T) {
		// Получаем заказ по номеру
		order, err := storage.GetOrderByNumber(ctx, "12345678903")
		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "12345678903", order.Number)

		// Проверяем несуществующий заказ
		order, err = storage.GetOrderByNumber(ctx, "99999999999")
		require.NoError(t, err)
		assert.Nil(t, order)
	})

	t.Run("GetOrdersByUserID", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Создаем еще один заказ
		_, err = storage.CreateOrder(ctx, user.ID, "98765432109")
		require.NoError(t, err)

		// Получаем все заказы пользователя
		orders, err := storage.GetOrdersByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, orders, 2)

		// Проверяем сортировку (от новых к старым)
		assert.Equal(t, "98765432109", orders[0].Number)
		assert.Equal(t, "12345678903", orders[1].Number)
	})

	t.Run("GetOrdersByStatus", func(t *testing.T) {
		// Получаем заказы со статусом NEW
		orders, err := storage.GetOrdersByStatus(ctx, []string{"NEW"})
		require.NoError(t, err)
		assert.Len(t, orders, 2)

		// Получаем заказы со статусом PROCESSING
		orders, err = storage.GetOrdersByStatus(ctx, []string{"PROCESSING"})
		require.NoError(t, err)
		assert.Len(t, orders, 0)
	})

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		// Обновляем статус заказа
		accrual := 100.0
		err := storage.UpdateOrderStatus(ctx, "12345678903", "PROCESSED", &accrual)
		require.NoError(t, err)

		// Проверяем, что статус обновился
		order, err := storage.GetOrderByNumber(ctx, "12345678903")
		require.NoError(t, err)
		assert.Equal(t, "PROCESSED", order.Status)
		assert.Equal(t, &accrual, order.Accrual)
	})

	t.Run("GetBalance", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Получаем баланс (должен быть создан автоматически)
		balance, err := storage.GetBalance(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, balance.UserID)
		assert.Equal(t, 0.0, balance.Current)
		assert.Equal(t, 0.0, balance.Withdrawn)
	})

	t.Run("UpdateBalance", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Обновляем баланс
		err = storage.UpdateBalance(ctx, user.ID, 500.0, 100.0)
		require.NoError(t, err)

		// Проверяем обновление
		balance, err := storage.GetBalance(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 500.0, balance.Current)
		assert.Equal(t, 100.0, balance.Withdrawn)
	})

	t.Run("CreateWithdrawal", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Создаем списание
		withdrawal, err := storage.CreateWithdrawal(ctx, user.ID, "testorder123", 50.0)
		require.NoError(t, err)
		assert.NotZero(t, withdrawal.ID)
		assert.Equal(t, user.ID, withdrawal.UserID)
		assert.Equal(t, "testorder123", withdrawal.Order)
		assert.Equal(t, 50.0, withdrawal.Sum)
	})

	t.Run("GetWithdrawalsByUserID", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Создаем еще одно списание
		_, err = storage.CreateWithdrawal(ctx, user.ID, "testorder456", 25.0)
		require.NoError(t, err)

		// Получаем все списания пользователя
		withdrawals, err := storage.GetWithdrawalsByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, withdrawals, 2)

		// Проверяем сортировку (от новых к старым)
		assert.Equal(t, "testorder456", withdrawals[0].Order)
		assert.Equal(t, "testorder123", withdrawals[1].Order)
	})

	t.Run("UpdateOrderStatusAndBalance", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Создаем заказ для тестирования с уникальным номером
		order, err := storage.CreateOrder(ctx, user.ID, "12345678904")
		require.NoError(t, err)

		// Устанавливаем начальный баланс
		err = storage.UpdateBalance(ctx, user.ID, 100.0, 0.0)
		require.NoError(t, err)

		// Атомарное обновление статуса заказа и баланса
		accrual := 50.0
		err = storage.UpdateOrderStatusAndBalance(ctx, order.Number, "PROCESSED", &accrual, user.ID, 150.0, 0.0)
		require.NoError(t, err)

		// Статус заказа обновился
		updatedOrder, err := storage.GetOrderByNumber(ctx, order.Number)
		require.NoError(t, err)
		assert.Equal(t, "PROCESSED", updatedOrder.Status)
		assert.Equal(t, &accrual, updatedOrder.Accrual)

		// Баланс обновился
		balance, err := storage.GetBalance(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 150.0, balance.Current)
		assert.Equal(t, 0.0, balance.Withdrawn)
	})

	t.Run("UpdateOrderStatusAndBalance_WithoutAccrual", func(t *testing.T) {
		user, err := storage.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)

		// Создаем еще один заказ с уникальным номером
		order, err := storage.CreateOrder(ctx, user.ID, "98765432108")
		require.NoError(t, err)

		// Обновление без начисления (accrual = nil)
		err = storage.UpdateOrderStatusAndBalance(ctx, order.Number, "PROCESSED", nil, user.ID, 150.0, 0.0)
		require.NoError(t, err)

		// Статус заказа обновился
		updatedOrder, err := storage.GetOrderByNumber(ctx, order.Number)
		require.NoError(t, err)
		assert.Equal(t, "PROCESSED", updatedOrder.Status)
		assert.Nil(t, updatedOrder.Accrual)
	})
}

// TestDatabaseStorage_Concurrent тестирует конкурентный доступ к базе данных
func TestDatabaseStorage_Concurrent(t *testing.T) {
	if !dbAvailable {
		t.Skip("Database not available, skipping test")
	}
	storage, err := NewDatabaseStorage(context.Background(), testDatabaseURI)
	if err != nil {
		t.Skipf("Skipping database tests: failed to connect to database: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "concurrentuser", "password")
	require.NoError(t, err)

	// Конкурентное создание заказов
	t.Run("ConcurrentOrderCreation", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				orderNumber := fmt.Sprintf("concurrent%d", id)
				_, err := storage.CreateOrder(ctx, user.ID, orderNumber)
				assert.NoError(t, err)
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Все заказы созданы
		orders, err := storage.GetOrdersByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, orders, numGoroutines)
	})
}

// TestDatabaseStorage_Transaction тестирует транзакционность операций
func TestDatabaseStorage_Transaction(t *testing.T) {
	if !dbAvailable {
		t.Skip("Database not available, skipping test")
	}
	storage, err := NewDatabaseStorage(context.Background(), testDatabaseURI)
	if err != nil {
		t.Skipf("Skipping database tests: failed to connect to database: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "transactionuser", "password")
	require.NoError(t, err)

	t.Run("BalanceUpdateTransaction", func(t *testing.T) {
		// Обновляем баланс
		err := storage.UpdateBalance(ctx, user.ID, 1000.0, 0.0)
		require.NoError(t, err)

		// Создаем списание
		_, err = storage.CreateWithdrawal(ctx, user.ID, "transactionorder", 100.0)
		require.NoError(t, err)

		// Обновляем баланс после списания
		err = storage.UpdateBalance(ctx, user.ID, 900.0, 100.0)
		require.NoError(t, err)

		// Финальное состояние
		balance, err := storage.GetBalance(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 900.0, balance.Current)
		assert.Equal(t, 100.0, balance.Withdrawn)

		withdrawals, err := storage.GetWithdrawalsByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, withdrawals, 1)
		assert.Equal(t, "transactionorder", withdrawals[0].Order)
		assert.Equal(t, 100.0, withdrawals[0].Sum)
	})
}

// cleanupDatabase очищает базу данных перед тестами
func cleanupDatabase(t *testing.T, storage *DatabaseStorage) {
	ctx := context.Background()
	// Удаляем все данные из таблиц
	queries := []string{
		"DELETE FROM withdrawals",
		"DELETE FROM balances",
		"DELETE FROM orders",
		"DELETE FROM users",
	}
	for _, query := range queries {
		_, err := storage.pool.Exec(ctx, query)
		if err != nil {
			t.Logf("Failed to cleanup: %v", err)
		}
	}
}
