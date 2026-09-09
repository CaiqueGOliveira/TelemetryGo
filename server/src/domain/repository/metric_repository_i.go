package repository

import "github.com/CaiqueGOliveira/TelemetryGo/src/domain"

type MetricRepository interface {
	Save(metric *domain.Metric) error
	FindAll(userID string) []*domain.Metric
}