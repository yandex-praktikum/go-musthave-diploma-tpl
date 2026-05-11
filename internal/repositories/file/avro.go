//go:generate mockgen -source=avro.go -destination=mocks/mock_avro_file.go -package=mocks
package file

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/linkedin/goavro/v2"
)

const (
	ordersAvroSchema = `
{
  "type": "record",
  "name": "Order",
  "namespace": "io.gophermart",
  "fields": [
    {"name": "number", "type": "string"},
    {"name": "user_id", "type": "string"},
    {"name": "status", "type": "string"},
    {"name": "accrual", "type": "long"},
    {"name": "uploaded_at", "type": "string"},
    {"name": "updated_at", "type": "string"}
  ]
}
`

	defaultListSize = 100
)

type (
	AvroFileStorage struct {
		filePath string
		mu       sync.Mutex
		codec    *goavro.Codec
	}
)

func NewAvroFileStorage(filePath string) (PersistStorageInterface, error) {
	codec, err := goavro.NewCodec(ordersAvroSchema)
	if err != nil {
		return nil, fmt.Errorf("can't create Avro-codec: %v", err)
	}

	return &AvroFileStorage{
		filePath: filePath,
		codec:    codec,
	}, nil
}

func (m *AvroFileStorage) Restore() (models.OrdersList, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fh, err := os.OpenFile(m.filePath, os.O_RDONLY, 0o600)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer fh.Close()

	orders := make(models.OrdersList, 0, defaultListSize)

	ocfReader, err := goavro.NewOCFReader(fh)
	if err != nil {
		return orders, nil
	}

	for ocfReader.Scan() {
		datum, err := ocfReader.Read()
		if err != nil {
			return nil, err
		}

		order, err := avroToModel(datum)
		if err != nil {
			return nil, fmt.Errorf("error has been occurred while reading Avro-file: %v", err)
		}
		orders = append(orders, order)
	}

	if err := ocfReader.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (m *AvroFileStorage) Save(orders models.OrdersList) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fh, err := os.Create(m.filePath)
	if err != nil {
		return err
	}
	defer fh.Close()

	ocfWriter, err := goavro.NewOCFWriter(goavro.OCFConfig{
		W:               fh,
		Codec:           m.codec,
		CompressionName: "snappy",
	})
	if err != nil {
		return err
	}

	var dataToWrite []any
	for _, order := range orders {
		avroMap := modelToAvro(order)
		dataToWrite = append(dataToWrite, avroMap)
	}

	if len(dataToWrite) > 0 {
		if err := ocfWriter.Append(dataToWrite); err != nil {
			return err
		}
	}

	return nil
}

// Erase удаляет файл
func (m *AvroFileStorage) Erase() error {
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

// modelToAvro преобразует models.Order в map для Avro
func modelToAvro(o *models.Order) map[string]any {
	return map[string]any{
		"number":      o.ID,
		"user_id":     o.UserID,
		"status":      o.Status,
		"accrual":     int64(o.Accrual), // Accrual (int) -> long
		"uploaded_at": o.UploadedAt.Format(time.RFC3339),
		"updated_at":  o.UpdatedAt.Format(time.RFC3339),
	}
}

// avroToModel преобразует Avro запись в models.Order
func avroToModel(datum any) (*models.Order, error) {
	record, ok := datum.(map[string]any)
	if !ok {
		return nil, errors.New("wrong Avro-record format")
	}

	order := &models.Order{}

	// Строковые поля
	if v, ok := record["number"]; ok {
		order.ID, ok = v.(string)
		if !ok {
			return nil, errors.New("can't convert number field")
		}
	}
	if v, ok := record["user_id"]; ok {
		order.UserID, ok = v.(string)
		if !ok {
			return nil, errors.New("can't convert user_id field")
		}
	}
	if v, ok := record["status"]; ok {
		order.Status, ok = v.(string)
		if !ok {
			return nil, errors.New("can't convert status field")
		}
	}

	// Числовые поля (Accrual)
	if v, ok := record["accrual"]; ok {
		// Avro long может вернуться как int64
		if val, ok := v.(int64); ok {
			order.Accrual = models.Accrual(val)
		} else if val, ok := v.(int); ok {
			// На случай, если кодек вернет int
			order.Accrual = models.Accrual(val)
		}
	}

	// Временные метки
	if v, ok := record["uploaded_at"]; ok {
		if str, ok := v.(string); ok {
			t, err := time.Parse(time.RFC3339, str)
			if err == nil {
				order.UploadedAt = t
			}
		}
	}

	if v, ok := record["updated_at"]; ok {
		if str, ok := v.(string); ok {
			t, err := time.Parse(time.RFC3339, str)
			if err == nil {
				order.UpdatedAt = t
			}
		}
	}

	return order, nil
}
