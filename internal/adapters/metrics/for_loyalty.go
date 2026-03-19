package metrics

import (
	"gmart/internal/domain"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusLoyaltyMetrics struct {
	dbDuration       *prometheus.HistogramVec
	withdrawCounts   *prometheus.CounterVec
	withdrawalAmount prometheus.Histogram
}

func NewForLoyalty(reg prometheus.Registerer) *PrometheusLoyaltyMetrics {
	// Бакеты для БД (1мс - 5с)
	dbBuckets := []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5}
	// Бакеты для сумм списаний (в копейках: 10р, 50р, 100р, 500р, 1000р, 5000р)
	amountBuckets := []float64{1000, 5000, 10000, 50000, 100000, 500000}

	m := &PrometheusLoyaltyMetrics{
		dbDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "gmart",
			Subsystem: "loyalty",
			Name:      "db_operation_duration_seconds",
			Help:      "Duration of database operations in loyalty repository",
			Buckets:   dbBuckets,
		}, []string{"op"}),

		withdrawCounts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "loyalty",
			Name:      "withdrawals_total",
			Help:      "Total number of withdrawal attempts by result",
		}, []string{"status"}),

		withdrawalAmount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "gmart",
			Subsystem: "loyalty",
			Name:      "withdrawal_amount_cents",
			Help:      "Distribution of withdrawal amounts in cents",
			Buckets:   amountBuckets,
		}),
	}

	reg.MustRegister(m.dbDuration, m.withdrawCounts, m.withdrawalAmount)
	return m
}

func (m *PrometheusLoyaltyMetrics) ObserveDB(op OpType, d time.Duration) {
	m.dbDuration.WithLabelValues(op.String()).Observe(d.Seconds())
}

func (m *PrometheusLoyaltyMetrics) IncWithdrawal(status string) {
	m.withdrawCounts.WithLabelValues(status).Inc()
}

func (m *PrometheusLoyaltyMetrics) ObserveWithdrawalAmount(amount domain.Amount) {
	m.withdrawalAmount.Observe(float64(amount))
}
