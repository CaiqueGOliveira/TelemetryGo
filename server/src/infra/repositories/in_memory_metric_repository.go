package repositories

import (
	"sort"
	"sync"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
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

func (r *InMemoryMetricRepository) List(userID string, filter repository.MetricFilter, limit int, offset int) ([]*domain.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Metric, 0)
	for _, metric := range r.metrics {
		if metric.UserId != userID {
			continue
		}
		if filter.Status != "" && metric.Status.String() != filter.Status {
			continue
		}
		if filter.Name != "" && metric.Name != filter.Name {
			continue
		}
		if filter.Service != "" && metric.Service != filter.Service {
			continue
		}
		if filter.Start != nil && metric.Timestamp.Before(*filter.Start) {
			continue
		}
		if filter.End != nil && metric.Timestamp.After(*filter.End) {
			continue
		}
		result = append(result, metric)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	return paginate(result, limit, offset), nil
}

func (r *InMemoryMetricRepository) Delete(userID string, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, metric := range r.metrics {
		if metric.UserId == userID && metric.Id == id {
			r.metrics = append(r.metrics[:i], r.metrics[i+1:]...)
			return nil
		}
	}

	return repository.ErrNotFound
}