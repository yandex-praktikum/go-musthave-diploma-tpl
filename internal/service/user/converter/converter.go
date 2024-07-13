package converter

import (
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	updateInterface "github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/update/interface"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/repository/user/file"
	"strconv"
)

func MetricCounterToMetricJSON(mg *entity.MetricCounter) *file.MetricJSON {
	return &file.MetricJSON{
		ID:    mg.Name,
		MType: string(entity.Counter),
		Delta: &mg.Value,
	}
}

func MetricGaugeToMetricJSON(mg *entity.MetricGauge) *file.MetricJSON {
	return &file.MetricJSON{
		ID:    mg.Name,
		MType: string(entity.Gauge),
		Value: &mg.Value,
	}
}

func MetricJSONToMetricCounter(mj *file.MetricJSON) *entity.MetricCounter {
	return &entity.MetricCounter{
		Type:  entity.Counter,
		Name:  mj.ID,
		Value: *mj.Delta,
	}
}

func MetricJSONToMetricGauge(mj *file.MetricJSON) *entity.MetricGauge {
	return &entity.MetricGauge{
		Type:  entity.Gauge,
		Name:  mj.ID,
		Value: *mj.Value,
	}
}

func MetricsJSONGaugeToMetricFields(metrics entity.MetricJSON) entity.MetricFields {
	return entity.MetricFields{
		MetricType:  metrics.MType,
		MetricName:  metrics.ID,
		MetricValue: strconv.FormatFloat(*metrics.Value, 'f', -1, 64),
	}
}
func MetricsJSONCounterToMetricFields(metrics entity.MetricJSON) entity.MetricFields {
	return entity.MetricFields{
		MetricType:  metrics.MType,
		MetricName:  metrics.ID,
		MetricValue: strconv.Itoa(int(*metrics.Delta)),
	}
}

func MetricJSONToMetricValueDTO(metrics entity.MetricJSON) updateInterface.GetMetricValueDto {
	return updateInterface.GetMetricValueDto{
		Type: metrics.MType,
		Name: metrics.ID,
	}
}
