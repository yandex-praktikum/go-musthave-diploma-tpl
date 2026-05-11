package file_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/file"

	"github.com/stretchr/testify/require"
)

func TestAvroFileStorage_SaveAndRestore(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "orders.avro")

	storage, err := file.NewAvroFileStorage(filePath)
	require.NoError(t, err, "NewAvroFileStorage should not return error")

	// Округляем время до секунд, так как при RFC3339 строковом формате теряются наносекунды
	now := time.Now().Truncate(time.Second)
	updateTime := now.Add(1 * time.Hour)

	originalOrders := models.OrdersList{
		&models.Order{
			ID:         "12345",
			UserID:     "user_1",
			Status:     models.OrderNewStatus,
			Accrual:    500, // 5.00 руб
			UploadedAt: now,
			UpdatedAt:  updateTime,
		},
		&models.Order{
			ID:         "67890",
			UserID:     "user_2",
			Status:     models.OrderProcessedStatus,
			Accrual:    10050, // 100.50 руб
			UploadedAt: now,
			UpdatedAt:  updateTime,
		},
	}

	err = storage.Save(originalOrders)
	require.NoError(t, err, "Save should not return error")

	_, err = os.Stat(filePath)
	require.NoError(t, err, "file should exist")

	restored, err := storage.Restore()
	require.NoError(t, err, "Restore should not return error")
	require.NotNil(t, restored)
	require.Len(t, restored, 2, "should restore 2 orders")

	// Проверка содержимого
	require.Equal(t, originalOrders[0].ID, restored[0].ID)
	require.Equal(t, originalOrders[0].Status, restored[0].Status)
	require.Equal(t, originalOrders[0].Accrual, restored[0].Accrual)
	require.True(t, originalOrders[0].UploadedAt.Equal(restored[0].UploadedAt), "UploadedAt should be equal")
	require.True(t, originalOrders[0].UpdatedAt.Equal(restored[0].UpdatedAt), "UpdatedAt should be equal")

	// В Avro схеме поле user_id присутствует, поэтому оно должно восстановиться
	require.Equal(t, originalOrders[0].UserID, restored[0].UserID, "UserID should be restored from Avro file")

	require.Equal(t, originalOrders[1].ID, restored[1].ID)
	require.Equal(t, originalOrders[1].Status, restored[1].Status)
	require.Equal(t, originalOrders[1].Accrual, restored[1].Accrual)
	require.Equal(t, originalOrders[1].UserID, restored[1].UserID)
}

func TestAvroFileStorage_Restore_FileNotExists(t *testing.T) {
	storage, err := file.NewAvroFileStorage("/non/existent/path/orders.avro")
	require.NoError(t, err)

	data, err := storage.Restore()
	require.NoError(t, err)
	require.Nil(t, data)
}

func TestAvroFileStorage_Restore_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "empty_*.avro")
	require.NoError(t, err)
	defer tmpFile.Close()

	storage, err := file.NewAvroFileStorage(tmpFile.Name())
	require.NoError(t, err)

	data, err := storage.Restore()
	require.NoError(t, err)
	require.Empty(t, data, "Empty file should return empty data slice")
}

func TestAvroFileStorage_Restore_InvalidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "bad_*.avro")
	require.NoError(t, err)
	_, _ = tmpFile.WriteString("not an avro binary content, just garbage")
	tmpFile.Close()

	storage, err := file.NewAvroFileStorage(tmpFile.Name())
	require.NoError(t, err)

	data, err := storage.Restore()

	require.NoError(t, err, "Implementation treats invalid file format as empty storage")
	require.Empty(t, data, "Invalid file should result in empty data")
}

func TestAvroFileStorage_Erase(t *testing.T) {
	t.Run("successfully erases existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "orders.avro")

		err := os.WriteFile(filePath, []byte{0x00}, 0o600)
		require.NoError(t, err, "failed to create test file")

		storage, err := file.NewAvroFileStorage(filePath)
		require.NoError(t, err)

		err = storage.Erase()
		require.NoError(t, err, "Erase should not return error for existing file")

		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "File should be deleted")
	})

	t.Run("returns nil if file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "non_existent.avro")

		storage, err := file.NewAvroFileStorage(filePath)
		require.NoError(t, err)

		err = storage.Erase()
		require.NoError(t, err, "Erase should return nil if file does not exist")
	})
}
