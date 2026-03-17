package metrics

import (
	"bytes"
	"gmart/internal/domain"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewForOrders_Registration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)

	// Ожидаем регистрацию 5 метрик (dbDuration, uploadCounts, acquireAttempts, listSize, finalizedCounts)
	mockReg.EXPECT().
		MustRegister(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1)

	m := NewForOrders(mockReg)
	assert.NotNil(t, m)
}

func TestPrometheusOrdersMetrics_Methods(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewForOrders(reg)

	t.Run("ObserveDB", func(t *testing.T) {
		duration := 150 * time.Millisecond
		m.ObserveDB(OpQuery, duration)

		// Проверяем общее количество записей в гистограмме
		count := testutil.CollectAndCount(m.dbDuration)
		assert.Equal(t, 1, count)
	})

	t.Run("IncOrderUpload", func(t *testing.T) {
		m.IncOrderUpload("new")
		m.IncOrderUpload("new")
		m.IncOrderUpload("conflict")

		val := testutil.ToFloat64(m.uploadCounts.WithLabelValues("new"))
		assert.Equal(t, 2.0, val)

		valConflict := testutil.ToFloat64(m.uploadCounts.WithLabelValues("conflict"))
		assert.Equal(t, 1.0, valConflict)
	})

	t.Run("IncAcquireAttempt", func(t *testing.T) {
		m.IncAcquireAttempt(true)
		m.IncAcquireAttempt(false)

		valTrue := testutil.ToFloat64(m.acquireAttempts.WithLabelValues("true"))
		assert.Equal(t, 1.0, valTrue)
	})

	t.Run("ObserveListSize", func(t *testing.T) {
		m.ObserveListSize(42)

		// Проверяем попадание в бакеты через текстовое сравнение
		err := testutil.CollectAndCompare(m.listSize, bytes.NewBufferString(`
			# HELP gmart_orders_list_size_rows Number of orders returned in a single list request
			# TYPE gmart_orders_list_size_rows histogram
			gmart_orders_list_size_rows_bucket{le="0"} 0
			gmart_orders_list_size_rows_bucket{le="5"} 0
			gmart_orders_list_size_rows_bucket{le="10"} 0
			gmart_orders_list_size_rows_bucket{le="20"} 0
			gmart_orders_list_size_rows_bucket{le="50"} 1
			gmart_orders_list_size_rows_bucket{le="100"} 1
			gmart_orders_list_size_rows_bucket{le="500"} 1
			gmart_orders_list_size_rows_bucket{le="+Inf"} 1
			gmart_orders_list_size_rows_sum 42
			gmart_orders_list_size_rows_count 1
		`), "gmart_orders_list_size_rows")
		assert.NoError(t, err)
	})

	t.Run("IncOrderFinalized", func(t *testing.T) {
		status := domain.OrderStatus("PROCESSED")
		m.IncOrderFinalized(status)

		val := testutil.ToFloat64(m.finalizedCounts.WithLabelValues("PROCESSED"))
		assert.Equal(t, 1.0, val)
	})
}
