package converter

import (
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
)

func MetricJSONToMetricCounter(metricJSON *entity.MetricJSON) *entity.MetricCounter {
	return &entity.MetricCounter{
		Type:  entity.Counter,
		Name:  metricJSON.ID,
		Value: *metricJSON.Delta,
	}
}

func MetricJSONToMetricGauge(metricJSON *entity.MetricJSON) *entity.MetricGauge {
	return &entity.MetricGauge{
		Type:  entity.Gauge,
		Name:  metricJSON.ID,
		Value: *metricJSON.Value,
	}
}
