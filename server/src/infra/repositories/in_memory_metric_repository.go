package repositories

import (
	"sync"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
)

type InMemoryMetricRepository struct {
	mu      sync.Mutex
	metrics []*domain.Metric
}

func NewInMemoryMetricRepository() *InMemoryMetricRepository {
	return &InMemoryMetricRepository{}
}

func (r *InMemoryMetricRepository) Save(metric *domain.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics = append(r.metrics, metric)
	return nil
}

func (r *InMemoryMetricRepository) FindAll(userID string) []*domain.Metric {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Metric, 0, len(r.metrics))
	for _, metric := range r.metrics {
		if metric.UserId == userID {
			result = append(result, metric)
		}
	}
	return result
}
