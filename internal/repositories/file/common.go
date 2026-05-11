//go:generate mockgen -source=common.go -destination=mocks/mock_common_file.go -package=mocks
package file

import (
	"github.com/KalessinD/gophermart/internal/models"
)

const (
	JSONTYpe = "json"
	AvroType = "avro"
)

type (
	PersistStorageInterface interface {
		Save(orders models.OrdersList) error
		Restore() (models.OrdersList, error)
		Erase() error
	}
)

/*
Фабрика.
Возвращает объект файлового хранилища согласно переданому типу.
Тип может быть одним из двух:

- json - file.JSONTYpe
- avro - file.AvroType
*/
func NewFileStorage(storageType, filePath string) (PersistStorageInterface, error) {
	switch storageType {
	case JSONTYpe:
		return NewJSONFileStorage(filePath)
	case AvroType:
	default:
		return NewAvroFileStorage(filePath)
	}
	return NewAvroFileStorage(filePath)
}
