//go:generate mockgen -source=json.go -destination=mocks/mock_json_file.go -package=mocks
package file

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/KalessinD/gophermart/internal/models"
)

type (
	JSONFileStorage struct {
		filePath string
		mu       sync.Mutex
	}
)

func NewJSONFileStorage(filePath string) (PersistStorageInterface, error) {
	dir := filepath.Dir(filePath)

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", dir)
		}
		return nil, fmt.Errorf("unable to access directory '%s': %w", dir, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dir)
	}

	return &JSONFileStorage{
		filePath: filePath,
	}, nil
}

func (m *JSONFileStorage) Restore() (models.OrdersList, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fh, err := os.OpenFile(m.filePath, os.O_RDONLY, 0o600)
	if err != nil {
		// if not exists, it's not an error
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer fh.Close()

	// reading the small data
	data, err := io.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	// делать нечего, пойдём курить бамбук
	if len(data) == 0 {
		return nil, nil
	}

	var orders models.OrdersList
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (m *JSONFileStorage) Save(orders models.OrdersList) error {
	data, err := json.Marshal(orders)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// os.WriteFile открывает файл с флагом O_TRUNC
	return os.WriteFile(m.filePath, data, 0o600)
}

// Erase удаляет файл с сохраненными данными.
func (m *JSONFileStorage) Erase() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := os.Remove(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}
