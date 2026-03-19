package metrics

import (
	"gmart/internal/domain"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusOrdersMetrics struct {
	dbDuration      *prometheus.HistogramVec
	uploadCounts    *prometheus.CounterVec
	acquireAttempts *prometheus.CounterVec
	listSize        prometheus.Histogram
	finalizedCounts *prometheus.CounterVec
}

func NewForOrders(reg prometheus.Registerer) *PrometheusOrdersMetrics {
	dbBuckets := []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5}
	sizeBuckets := []float64{0, 5, 10, 20, 50, 100, 500}

	m := &PrometheusOrdersMetrics{
		dbDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "gmart",
			Subsystem: "orders",
			Name:      "db_operation_duration_seconds",
			Help:      "Duration of database operations",
			Buckets:   dbBuckets,
		}, []string{"op"}),

		uploadCounts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "orders",
			Name:      "upload_total",
			Help:      "Total number of order upload attempts",
		}, []string{"result"}),

		acquireAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "orders",
			Name:      "acquire_attempts_total",
			Help:      "Total number of worker acquire attempts",
		}, []string{"found"}),

		listSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "gmart",
			Subsystem: "orders",
			Name:      "list_size_rows",
			Help:      "Number of orders returned in a single list request",
			Buckets:   sizeBuckets,
		}),

		finalizedCounts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "orders",
			Name:      "finalized_total",
			Help:      "Total number of orders that reached final status",
		}, []string{"status"}),
	}

	reg.MustRegister(m.dbDuration, m.uploadCounts, m.acquireAttempts, m.listSize, m.finalizedCounts)
	return m
}

// ObserveDB записывает время выполнения запроса к БД
func (m *PrometheusOrdersMetrics) ObserveDB(op OpType, d time.Duration) {
	m.dbDuration.WithLabelValues(op.String()).Observe(d.Seconds())
}

// IncOrderUpload инкрементирует счетчик загрузок с результатом (new, conflict, exists)
func (m *PrometheusOrdersMetrics) IncOrderUpload(result string) {
	m.uploadCounts.WithLabelValues(result).Inc()
}

// IncAcquireAttempt
func (m *PrometheusOrdersMetrics) IncAcquireAttempt(found bool) {
	label := "false"
	if found {
		label = "true"
	}
	m.acquireAttempts.WithLabelValues(label).Inc()
}

// ObserveListSize записывает количество найденных заказов
func (m *PrometheusOrdersMetrics) ObserveListSize(size int) {
	m.listSize.Observe(float64(size))
}

// IncOrderFinalized
func (m *PrometheusOrdersMetrics) IncOrderFinalized(status domain.OrderStatus) {
	m.finalizedCounts.WithLabelValues(status.String()).Inc()
}
