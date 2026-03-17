package metrics

import (
	"bytes"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewForAuth_Registration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)

	// MustRegister принимает variadic (...Collector).
	// Мы ожидаем регистрацию двух объектов (Counter и Histogram).
	mockReg.EXPECT().
		MustRegister(gomock.Any(), gomock.Any()).
		Times(1)

	m := NewForAuth(mockReg)
	assert.NotNil(t, m)
}

func TestPrometheusAuthMetrics_Methods(t *testing.T) {
	// Используем реальный Registry для проверки логики записи
	reg := prometheus.NewRegistry()
	m := NewForAuth(reg)

	t.Run("IncLoginError", func(t *testing.T) {
		reason := "user_not_found"
		m.IncLoginError(reason)
		m.IncLoginError(reason)

		// Проверяем инкремент счетчика
		val := testutil.ToFloat64(m.loginErrors.WithLabelValues(reason))
		assert.Equal(t, 2.0, val)
	})

	t.Run("ObserveBcrypt", func(t *testing.T) {
		duration := 350 * time.Millisecond
		m.ObserveBcrypt(duration)

		// 1. Проверяем, что в гистограмме появилось наблюдение.
		// CollectAndCount принимает Collector (всю гистограмму m.bcrypt).
		count := testutil.CollectAndCount(m.bcrypt)
		assert.Equal(t, 1, count)

		// 2. Чтобы проверить конкретное значение (sum) с учетом лейбла,
		// используем GatherAndCompare или специфичный хелпер для Lint/Сравнения,
		// но для юнит-теста достаточно убедиться, что данные попали в реестр.
		err := testutil.CollectAndCompare(m.bcrypt, bytes.NewBufferString(`
			# HELP gmart_auth_bcrypt_duration_seconds Duration of bcrypt operations
			# TYPE gmart_auth_bcrypt_duration_seconds histogram
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.05"} 0
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.1"} 0
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.2"} 0
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.3"} 0
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.4"} 1
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.5"} 1
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="0.75"} 1
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="1"} 1
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="2"} 1
			gmart_auth_bcrypt_duration_seconds_bucket{type="bcrypt",le="+Inf"} 1
			gmart_auth_bcrypt_duration_seconds_sum{type="bcrypt"} 0.35
			gmart_auth_bcrypt_duration_seconds_count{type="bcrypt"} 1
		`), "gmart_auth_bcrypt_duration_seconds")

		assert.NoError(t, err)
	})
}
