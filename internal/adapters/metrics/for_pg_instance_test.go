package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewForPgInstance_Registration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)
	mockReg.EXPECT().
		MustRegister(
			gomock.Any(), // status
			gomock.Any(), // offline
			gomock.Any(), // retries
			gomock.Any(), // latency
		).
		Times(1)

	m := NewForPgInstance(mockReg)
	assert.NotNil(t, m)
}

func TestPrometheusPgInstanceMetrics_Methods(t *testing.T) {
	// Создаем реальный реестр для проверки значений метрик
	reg := prometheus.NewRegistry()
	m := NewForPgInstance(reg)

	const inst = "main_db"

	t.Run("SetStatus", func(t *testing.T) {
		m.SetStatus(inst, true)
		assert.Equal(t, 1.0, testutil.ToFloat64(m.status.WithLabelValues(inst)))

		m.SetStatus(inst, false)
		assert.Equal(t, 0.0, testutil.ToFloat64(m.status.WithLabelValues(inst)))
	})

	t.Run("IncOfflineEvent", func(t *testing.T) {
		m.IncOfflineEvent(inst)
		m.IncOfflineEvent(inst)
		assert.Equal(t, 2.0, testutil.ToFloat64(m.offline.WithLabelValues(inst)))
	})

	t.Run("IncRetry", func(t *testing.T) {
		m.IncRetry(inst, OpQuery)
		assert.Equal(t, 1.0, testutil.ToFloat64(m.retries.WithLabelValues(inst, "query")))
	})

	t.Run("ObserveLatency", func(t *testing.T) {
		m.ObserveLatency(inst, OpExec, 0.5)
		// testutil.CollectAndCount проверяет наличие данных в гистограмме
		count := testutil.CollectAndCount(m.latency)
		assert.Equal(t, 1, count)
	})
}
