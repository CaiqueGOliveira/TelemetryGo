package repository

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/google/uuid"
)

type MetricRepository interface {
	Save(metric *domain.Metric) error
	List(userID string, filter MetricFilter, limit int, offset int) ([]*domain.Metric, error)
	Delete(userID string, id uuid.UUID) error
}
