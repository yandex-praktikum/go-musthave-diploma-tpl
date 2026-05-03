package file_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONStorage_SaveAndRestore(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "orders.json")
	storage, _ := file.NewJSONFileStorage(filePath)

	// Подготовка данных.
	// Важно: time.Time нужно округлять, так как JSON маршалинг может потерять наносекунды
	now := time.Now().Truncate(time.Second)

	originalOrders := models.OrdersList{
		&models.Order{
			ID:         "12345",
			UserID:     "user_1", // UserID не запишется в файл (json:"-")
			Status:     models.OrderNewStatus,
			Accrual:    500,
			UploadedAt: now,
		},
		&models.Order{
			ID:         "67890",
			UserID:     "user_2",
			Status:     models.OrderProcessedStatus,
			Accrual:    100,
			UploadedAt: now,
		},
	}

	// Тест сохранения
	err := storage.Save(originalOrders)
	require.NoError(t, err, "Save should not return error")

	_, err = os.Stat(filePath)
	require.NoError(t, err, "file should exist")

	// Тест восстановления
	restored, err := storage.Restore()
	require.NoError(t, err, "Restore should not return error")
	require.NotNil(t, restored)
	require.Len(t, restored, 2, "should restore 2 orders")

	// Проверка содержимого
	// Порядок в JSON массиве должен сохраниться
	require.Equal(t, originalOrders[0].ID, restored[0].ID)
	require.Equal(t, originalOrders[0].Status, restored[0].Status)
	require.Equal(t, originalOrders[0].Accrual, restored[0].Accrual)
	require.Equal(t, originalOrders[0].UploadedAt, restored[0].UploadedAt)

	// UserID имеет тег `json:"-"`, поэтому при восстановлении он будет пустым
	// нам это неважно, т.к. accrual принимает только номер заказа, а в БД запись мы обновляем по orderID
	require.Empty(t, restored[0].UserID, "UserID should not be restored from file")

	require.Equal(t, originalOrders[1].ID, restored[1].ID)
	require.Equal(t, originalOrders[1].Status, restored[1].Status)
}

func TestFileStorage_Restore_FileNotExists(t *testing.T) {
	_, err := file.NewJSONFileStorage("/non/existent/path/orders.json")
	assert.Error(t, err, "expects error")
}

func TestFileStorage_Restore_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "empty_*.json")
	require.NoError(t, err)
	defer tmpFile.Close()

	storage, _ := file.NewJSONFileStorage(tmpFile.Name())

	data, err := storage.Restore()
	require.NoError(t, err)
	require.Nil(t, data, "Empty file should return nil data")
}

func TestFileStorage_Restore_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "bad_*.json")
	require.NoError(t, err)
	_, _ = tmpFile.WriteString("not a json")
	tmpFile.Close()

	storage, _ := file.NewJSONFileStorage(tmpFile.Name())

	_, err = storage.Restore()
	require.Error(t, err, "Should return error on invalid JSON")

	// Проверяем, что ошибка именно о синтаксисе JSON
	_, ok := err.(*json.SyntaxError)
	require.True(t, ok, "Error should be JSON syntax error")
}

func TestFileStorage_Erase(t *testing.T) {
	t.Run("successfully erases existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "orders.json")

		err := os.WriteFile(filePath, []byte(`[{"id":"1"}]`), 0o600)
		require.NoError(t, err, "failed to create test file")

		storage, _ := file.NewJSONFileStorage(filePath)

		err = storage.Erase()
		require.NoError(t, err, "Erase should not return error for existing file")

		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "File should be deleted")
	})

	t.Run("returns nil if file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "non_existent.json")

		storage, _ := file.NewJSONFileStorage(filePath)

		err := storage.Erase()

		require.NoError(t, err, "Erase should return nil if file does not exist")
	})
}
