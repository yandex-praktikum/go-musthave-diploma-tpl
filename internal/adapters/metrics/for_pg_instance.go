package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusPgInstanceMetrics struct {
	status  *prometheus.GaugeVec
	offline *prometheus.CounterVec
	retries *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func NewForPgInstance(reg prometheus.Registerer) *PrometheusPgInstanceMetrics {
	// Кастомные бакеты для БД: от 1мс до 5сек.
	dbBuckets := []float64{.001, .002, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5}

	m := &PrometheusPgInstanceMetrics{
		status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "gmart",
			Subsystem: "postgresql",
			Name:      "pg_instance_ready",
			Help:      "1 if instance is online, 0 if offline",
		}, []string{"instance"}),

		offline: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "postgresql",
			Name:      "pg_instance_offline_total",
			Help:      "Total number of offline events",
		}, []string{"instance"}),

		retries: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gmart",
			Subsystem: "postgresql",
			Name:      "pg_instance_retries_total",
			Help:      "Total number of retries",
		}, []string{"instance", "type"}),

		latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "gmart",
			Subsystem: "postgresql",
			Name:      "pg_instance_duration_seconds",
			Help:      "Duration of operations",
			Buckets:   dbBuckets, // Используем наши бакеты
		}, []string{"instance", "type"}),
	}

	// Без метрик не стартуем
	reg.MustRegister(m.status, m.offline, m.retries, m.latency)

	return m
}

func (m *PrometheusPgInstanceMetrics) SetStatus(inst string, online bool) {
	val := 0.0
	if online {
		val = 1.0
	}
	m.status.WithLabelValues(inst).Set(val)
}

func (m *PrometheusPgInstanceMetrics) IncOfflineEvent(inst string) {
	m.offline.WithLabelValues(inst).Inc()
}

func (m *PrometheusPgInstanceMetrics) IncRetry(inst string, opType OpType) {
	m.retries.WithLabelValues(inst, opType.String()).Inc()
}

func (m *PrometheusPgInstanceMetrics) ObserveLatency(inst string, opType OpType, d float64) {
	m.latency.WithLabelValues(inst, opType.String()).Observe(d)
}
