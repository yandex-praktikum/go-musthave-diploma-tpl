package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusAuthMetrics struct {
	loginErrors *prometheus.CounterVec
	bcrypt      *prometheus.HistogramVec
}

func NewForAuth(reg prometheus.Registerer) *PrometheusAuthMetrics {
	// Бакеты для Bcrypt (обычно от 50мс до 1сек)
	bcryptBuckets := []float64{.05, .1, .2, .3, .4, .5, .75, 1, 2}

	m := &PrometheusAuthMetrics{
		loginErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "auth",
			Name:      "login_errors_total",
			Help:      "Total number of login failures by reason",
		}, []string{"reason"}),

		bcrypt: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "gmart",
			Subsystem: "auth",
			Name:      "bcrypt_duration_seconds",
			Help:      "Duration of bcrypt operations",
			Buckets:   bcryptBuckets,
		}, []string{"type"}),
	}

	// Без метрик не стартуем
	reg.MustRegister(m.loginErrors, m.bcrypt)

	return m
}

// IncLoginError фиксирует причину отказа в доступе
func (m *PrometheusAuthMetrics) IncLoginError(reason string) {
	m.loginErrors.WithLabelValues(reason).Inc()
}

// ObserveBcrypt записывает время работы хеш-функции
func (m *PrometheusAuthMetrics) ObserveBcrypt(d time.Duration) {
	m.bcrypt.WithLabelValues(OpBcrypt.String()).Observe(d.Seconds())
}
