package dtos

import "github.com/CaiqueGOliveira/TelemetryGo/src/domain"

type MetricResponseDto struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	Service   string  `json:"service"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

func ToMetricResponseDto(metric *domain.Metric) MetricResponseDto {
	return MetricResponseDto{
		Id:        metric.Id.String(),
		Name:      metric.Name,
		Service:   metric.Service,
		Value:     metric.Value,
		Unit:      metric.Unit,
		Status:    metric.Status.String(),
		Timestamp: metric.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToMetricResponseDtos(metrics []*domain.Metric) []MetricResponseDto {
	result := make([]MetricResponseDto, 0, len(metrics))
	for _, metric := range metrics {
		result = append(result, ToMetricResponseDto(metric))
	}
	return result
}
