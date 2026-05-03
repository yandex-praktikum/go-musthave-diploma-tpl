package gophermart_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgMigrator_Apply(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	migrator := gophermart.NewPgMigrator(db)

	tmpDir := t.TempDir()

	t.Run("successful migration", func(t *testing.T) {
		// Создаем временный файл с SQL
		fileName := "001_init.sql"
		filePath := filepath.Join(tmpDir, fileName)
		sqlContent := []byte("CREATE TABLE users (id INT);")
		err := os.WriteFile(filePath, sqlContent, 0o600)
		require.NoError(t, err)

		// Ожидаем выполнение именно этого SQL
		mock.ExpectExec("CREATE TABLE users").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = migrator.Apply(context.Background(), tmpDir, []string{filePath})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("file does not exist", func(t *testing.T) {
		// Путь к несуществующему файлу
		nonExistentFileName := "non_existent.sql"
		nonExistentPath := filepath.Join(tmpDir, nonExistentFileName)

		// Согласно логике: if os.IsNotExist(err) { continue }
		// Ошибки быть не должно, и DB вызываться не должна
		err := migrator.Apply(context.Background(), tmpDir, []string{nonExistentPath})
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sql execution error", func(t *testing.T) {
		fileName := "002_init.sql"
		filePath := filepath.Join(tmpDir, fileName)
		sqlContent := []byte("BAD SQL SYNTAX")
		err := os.WriteFile(filePath, sqlContent, 0o600)
		require.NoError(t, err)

		// Эмулируем ошибку БД
		mock.ExpectExec("BAD SQL SYNTAX").
			WillReturnError(errors.New("syntax error"))

		err = migrator.Apply(context.Background(), tmpDir, []string{filePath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "syntax error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("read file error (permission denied)", func(t *testing.T) {
		// Создаем директорию без прав на чтение/запись файлов,
		// или файл без прав чтения, чтобы вызвать ошибку os.ReadFile
		// (Пропускаем на Windows, так как права работают иначе)
		if os.Getenv("GOOS") == "windows" {
			t.Skip("skipping on windows")
		}

		fileName := "protected"
		protectedDir := filepath.Join(tmpDir, fileName)
		err := os.Mkdir(protectedDir, 0o000) // никаких прав
		require.NoError(t, err)

		// Восстанавливаем права для очистки
		defer func() { _ = os.Chmod(protectedDir, 0o755) }()

		// os.ReadFile упадет с permission denied
		err = migrator.Apply(context.Background(), tmpDir, []string{protectedDir})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("read directory instead of file", func(t *testing.T) {
		// Попытка "прочитать" директорию как файл вызовет ошибку (не IsNotExist)
		dirName := "somedir"
		dirPath := filepath.Join(tmpDir, dirName)
		err := os.Mkdir(dirPath, 0o755)
		require.NoError(t, err)

		err = migrator.Apply(context.Background(), tmpDir, []string{dirPath})
		require.Error(t, err)
		// Ошибка будет "read /path: is a directory"
		assert.Contains(t, err.Error(), "is a directory")
	})
}
