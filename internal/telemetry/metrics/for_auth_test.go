package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewForAuth_Registration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)

	// Ожидаем один вызов MustRegister с двумя аргументами (loginErrors и bcrypt)
	mockReg.EXPECT().
		MustRegister(gomock.Any(), gomock.Any()).
		Times(1)

	m := NewForAuth(mockReg)
	assert.NotNil(t, m)
}

func TestPrometheusAuthMetrics_Methods(t *testing.T) {
	// Используем реальный Registry для проверки записи значений
	reg := prometheus.NewRegistry()
	m := NewForAuth(reg)

	t.Run("IncLoginError", func(t *testing.T) {
		reason := "user_not_found"
		m.IncLoginError(reason)
		m.IncLoginError(reason)

		// Проверяем, что счетчик увеличился до 2
		val := testutil.ToFloat64(m.loginErrors.WithLabelValues(reason))
		assert.Equal(t, 2.0, val)
	})

	t.Run("ObserveBcrypt", func(t *testing.T) {
		duration := 0.350 // 350ms
		m.ObserveBcrypt(duration)

		// Проверяем наличие данных в гистограмме
		// CollectAndCount возвращает общее количество наблюдений
		count := testutil.CollectAndCount(m.bcrypt)
		assert.Equal(t, 1, count)
	})
}
